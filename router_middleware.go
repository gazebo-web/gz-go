package ign

import (
	"encoding/json"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"github.com/jpillora/go-ogle-analytics"
	"github.com/mssola/user_agent"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Handler represents an HTTP Handler that can also return a ErrMsg
// See https://blog.golang.org/error-handling-and-go
type Handler func(*gorm.DB, http.ResponseWriter, *http.Request) *ErrMsg

// HandlerWithResult represents an HTTP Handler that that has a result
type HandlerWithResult func(tx *gorm.DB, w http.ResponseWriter,
	r *http.Request) (interface{}, *ErrMsg)

// TypeJSONResult represents a function result that can be exported to JSON
type TypeJSONResult struct {
	wrapperField string
	fn           HandlerWithResult
	wrapWithTx   bool
}

// ProtoResult provides protobuf serialization for handler results
type ProtoResult HandlerWithResult

// JSONResult provides JSON serialization for handler results
func JSONResult(handler HandlerWithResult) TypeJSONResult {
	return TypeJSONResult{"", handler, true}
}

// IsBotHandler decides which handler to use whether the request was made by a
// bot or a user.
func IsBotHandler(botHandler http.Handler, userHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var handler http.Handler
		ua := user_agent.New(r.Header.Get("User-Agent"))
		if (ua.Bot()) {
			handler = botHandler
		} else {
			handler = userHandler
		}
		handler.ServeHTTP(w, r)
	})
}

// JSONResultNoTx provides JSON serialization for handler results
func JSONResultNoTx(handler HandlerWithResult) TypeJSONResult {
	return TypeJSONResult{"", handler, false}
}

// JSONListResult provides JSON serialization for handler results that are
// slices of objects.
func JSONListResult(wrapper string, handler HandlerWithResult) TypeJSONResult {
	return TypeJSONResult{wrapper, handler, true}
}

/////////////////////////////////////////////////

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	txFunc := dbTransactionWrapper(handlerToHandlerWithResult(fn))
	if _, err := txFunc(w, r); err != nil {
		reportJSONError(w, r, *err)
	}
}

/////////////////////////////////////////////////

// basicHandlerWith represents a basic handler function that returns a result and an error.
type basicHandlerWithResult func(w http.ResponseWriter, r *http.Request) (interface{}, *ErrMsg)

// IsSQLTxError checks if the given error is a sqlTx error.
// Note: we need to do that by testing its error message.
func IsSQLTxError(err error) bool {
	return err != nil && strings.ToLower(err.Error()) == "sql: transaction has already been committed or rolled back"
}

// dbTransactionWrapper handles opening and closing of a DB Transaction.
// It invokes the given handler with the created TX.
// By using this wrapper , real handlers won't need to open and close the TX.
// IMPORTANT NOTE: note that once you write data (not headers) into the
// ResponseWriter then the status code is set to 200 (OK). Keep that in mind
// when coding your Handler logic (eg. when using fmt.Fprint(w, ...))
func dbTransactionWrapper(handler HandlerWithResult) basicHandlerWithResult {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, *ErrMsg) {
		tx := gServer.Db.Begin()
		if tx.Error != nil {
			return nil, NewErrorMessageWithBase(ErrorNoDatabase, tx.Error)
		}

		defer func() {
			// check for panic (to close sql connections)
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p) // re-throw panic after Rollback
			}
		}()
		result, em := handler(tx, w, r)
		if em != nil {
			tx.Rollback()
		} else {
			// Commit DB transaction
			err := tx.Commit().Error
			if err != nil && !IsSQLTxError(err) {
				// re-throw error if different than TX already committed/rollbacked err
				result, em = nil, NewErrorMessageWithBase(ErrorNoDatabase, err)
			}
		}
		return result, em
	}
}

// handlerToHandlerWithResult converts an ign.Handler to an
// ign.HandlerWithResult.
func handlerToHandlerWithResult(handler Handler) HandlerWithResult {
	return func(tx *gorm.DB, w http.ResponseWriter, r *http.Request) (interface{}, *ErrMsg) {
		err := handler(tx, w, r)
		return nil, err
	}
}

func (t TypeJSONResult) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var txFunc basicHandlerWithResult
	if t.wrapWithTx {
		txFunc = dbTransactionWrapper(t.fn)
	} else {
		txFunc = func(w http.ResponseWriter, r *http.Request) (interface{}, *ErrMsg) {
			return t.fn(gServer.Db, w, r)
		}
	}
	result, err := txFunc(w, r)
	if err != nil {
		reportJSONError(w, r, *err)
		return
	}

	var data interface{}
	// Is there any wrapper field to cut off ?
	if t.wrapperField != "" {
		value := reflect.ValueOf(result)
		fieldValue := reflect.Indirect(value).FieldByName(t.wrapperField)
		data = fieldValue.Interface()
		// If the underlying data is an empty slice then force the creation of
		// an empty json `[]` as output
		if fieldValue.Kind() == reflect.Slice && fieldValue.Len() == 0 {
			data = make([]string, 0)
		}
	} else {
		data = result
	}
	w.Header().Set("Content-Type", "application/json")
	// Marshal the response into a JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		em := NewErrorMessageWithBase(ErrorMarshalJSON, err)
		reportJSONError(w, r, *em)
		return
	}
}

/////////////////////////////////////////////////

func (fn ProtoResult) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	txFunc := dbTransactionWrapper(HandlerWithResult(fn))
	result, err := txFunc(w, r)
	if err != nil {
		reportJSONError(w, r, *err)
		return
	}

	// Marshal the protobuf data and write it out.
	var pm = result.(proto.Message)
	data, e := proto.Marshal(pm)
	if e != nil {
		em := NewErrorMessageWithBase(ErrorMarshalProto, e)
		reportJSONError(w, r, *em)
		return
	}
	w.Header().Set("Content-Type", "application/arraybuffer")
	w.Write(data)
}

/////////////////////////////////////////////////

// ReportJSONError logs an error message and return an HTTP error including
// JSON payload
func reportJSONError(w http.ResponseWriter, r *http.Request, errMsg ErrMsg) {
	errMsg.UserAgent = r.UserAgent()
	errMsg.RemoteAddress = getIPAddress(r)
	if errMsg.Route == "" {
		errMsg.Route = r.Method + " " + r.RequestURI
	}
	// Report the error to rollbar, and output to console
	LoggerFromRequest(r).Error(errMsg, r)

	output, err := json.Marshal(errMsg)
	if err != nil {
		reportError(w, "Unable to marshal JSON", http.StatusServiceUnavailable)
		return
	}

	http.Error(w, string(output), errMsg.StatusCode)
}

// reportError logs an error message and return an HTTP error
func reportError(w http.ResponseWriter, msg string, errCode int) {
	log.Println("Error in [" + Trace(3) + "]\n\t" + msg)
	http.Error(w, msg, errCode)
}

/////////////////////////////////////////////////

// JWTMiddlewareIgn wraps jwtmiddleware.JWTMiddleware so that we can create
// a custom AccessTokenHandler that first checks for a Private-Token and then
// checks for a JWT if the Private-Token doesn't exist.
type JWTMiddlewareIgn struct {
	*jwtmiddleware.JWTMiddleware
}

// AccessTokenHandler first checks for a Private-Token and then
// checks for a JWT if the Private-Token doesn't exist.
func (m *JWTMiddlewareIgn) AccessTokenHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Check if a Private-Token is used, which will supercede a JWT token.
	if token := r.Header.Get("Private-Token"); len(token) > 0 {

		var errorMsg *ErrMsg

		tx := gServer.UsersDb.Begin()
		defer func() {
			// check for panic (to close sql connections)
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p) // re-throw panic after Rollback
			}
		}()

		var accessToken *AccessToken
		if tx.Error != nil {
			errorMsg = NewErrorMessageWithBase(ErrorNoDatabase, tx.Error)
		} else {
			accessToken, errorMsg = ValidateAccessToken(token, tx)
		}

		if errorMsg != nil {
			logger := NewLoggerWithRollbarVerbosity("AccessTokenHandler", gServer.LogToStd, gServer.LogVerbosity, gServer.RollbarLogVerbosity)
			logger.Error(errorMsg)
			m.Options.ErrorHandler(w, r, errorMsg.Msg)
			tx.Rollback()
			return
		}

		if accessToken.LastUsed == nil {
			accessToken.LastUsed = new(time.Time)
		}

		*(accessToken.LastUsed) = time.Now()
		tx.Save(accessToken)
		tx.Commit()

		next(w, r)
	} else {
		m.HandlerWithNext(w, r, next)
	}
}

// CreateJWTOptionalMiddleware creates and returns a middleware that
// allows requests with optional JWT tokens.
func CreateJWTOptionalMiddleware(s *Server) negroni.HandlerFunc {
	// See https://github.com/auth0/go-jwt-middleware
	opt := jwtmiddleware.New(
		jwtmiddleware.Options{
			Debug:               false,
			CredentialsOptional: true,
			SigningMethod:       jwt.SigningMethodRS256,
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return jwt.ParseRSAPublicKeyFromPEM([]byte(s.pemKeyString))
			},
		})
	return negroni.HandlerFunc(opt.HandlerWithNext)
}

// CreateJWTRequiredMiddleware creates and returns a middleware that
// rejects requests that do not have a JWT token.
func CreateJWTRequiredMiddleware(s *Server) negroni.HandlerFunc {
	req := &JWTMiddlewareIgn{jwtmiddleware.New(jwtmiddleware.Options{
		Debug:               false,
		SigningMethod:       jwt.SigningMethodRS256,
		CredentialsOptional: false,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return jwt.ParseRSAPublicKeyFromPEM([]byte(s.pemKeyString))
		},
	})}

	return negroni.HandlerFunc(req.AccessTokenHandler)
}

// Middleware to ensure the DB instance exists.
// By having this middleware, then any route handler can safely assume the DB
// is present.
func requireDBMiddleware(w http.ResponseWriter, r *http.Request,
	next http.HandlerFunc) {

	if gServer.Db == nil {
		errMsg := ErrorMessage(ErrorNoDatabase)
		reportJSONError(w, r, errMsg)
	} else {
		next(w, r)
	}
}

// addCORSheadersMiddleware adds CORS related headers to an http response.
func addCORSheadersMiddleware(w http.ResponseWriter, r *http.Request,
	next http.HandlerFunc) {
	addCORSheaders(w)
	next(w, r)
}

// addCORSheaders adds the required Access Control headers to the HTTP response
func addCORSheaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods",
		"GET, HEAD, POST, PUT, PATCH, DELETE")

	w.Header().Set("Access-Control-Allow-Credentials", "true")

	w.Header().Set("Access-Control-Allow-Headers",
		`Accept, Accept-Language, Content-Language, Origin,
                  Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token,
                  Authorization`)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Header().Set("Access-Control-Expose-Headers", "Link, X-Total-Count, X-Ign-Resource-Version")
}

// getRequestID gets the request's X-Request-ID header OR, if the header is empty,
// returns a generated UUID string.
func getRequestID(r *http.Request) string {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.NewV4().String()
	}
	return reqID
}

/////////////////////////////////////////////////
// logger creates a middleware used to output HTTP requests.
func logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := getRequestID(r)
		logger := NewLoggerWithRollbarVerbosity(reqID, gServer.LogToStd, gServer.LogVerbosity, gServer.RollbarLogVerbosity)
		logCtx := NewContextWithLogger(r.Context(), logger)

		logger.Info(fmt.Sprintf("Incoming req: %s %s %s",
			r.Method,
			r.RequestURI,
			name,
		))
		// run the server logic
		inner.ServeHTTP(w, r.WithContext(logCtx))
		// log output
		logger.Info(fmt.Sprintf("Finished req: %s %s %s %s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		))
	})
}

/////////////////////////////////////////////////
// newGaEventTracking creates a new middleware to send events to Google Analytics.
// Events will be automatically created using route information.
// This middleware requires IGN_GA_TRACKING_ID and IGN_GA_APP_NAME
// env vars.
func newGaEventTracking(routeName string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(w, r)

		// Track event with GA, if enabled
		if gServer.GaAppName == "" || gServer.GaTrackingID == "" {
			return
		}
		c, err := ga.NewClient(gServer.GaTrackingID)
		if err != nil {
			LoggerFromRequest(r).Error("Error creating GA client", err, r)
			return
		}
		c.DataSource(gServer.GaAppName)
		c.ApplicationName(gServer.GaAppName)
		cat := gServer.GaCategoryPrefix + routeName
		action := r.Method
		e := ga.NewEvent(cat, action).Label(r.URL.String())
		if err := c.Send(e); err != nil {
			fmt.Println("Error while sending event to GA", err)
		}
	}
}
