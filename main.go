package gz

// Import this file's dependencies
import (
	"errors"
	"fmt"
	"github.com/gazebo-web/gz-go/v7/monitoring"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rollbar/rollbar-go"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	// Needed by dbInit
	_ "github.com/go-sql-driver/mysql"
)

// Server encapsulates information needed by a downstream application
type Server struct {
	// / Global database interface
	Db *gorm.DB

	// / Global database to the user database interface
	UsersDb *gorm.DB

	Router *mux.Router

	// monitoring contains an optional monitoring provider used to
	// keep track of server metrics. The metrics are defined by the provider.
	// If set, the server will automatically add monitoring middleware when
	// configuring routes and expose an endpoint to allow the monitoring system
	// to scrape metric data.
	monitoring monitoring.Provider

	// Port used for non-secure requests
	HTTPPort string

	// SSLport used for secure requests
	SSLport string

	// SSLCert is the path to the SSL certificate.
	SSLCert string

	// SSLKey is the path to the SSL private key.
	SSLKey string

	// DbConfig contains information about the database
	DbConfig DatabaseConfig

	// IsTest is true when tests are running.
	IsTest bool

	// / Auth0 public key used for token validation
	auth0RsaPublickey string
	// PEM Key string generated from the auth0RsaPublickey value
	pemKeyString string

	// Google Analytics tracking ID. The format is UA-XXXX-Y
	GaTrackingID string

	// Google Analytics Application Name
	GaAppName string

	// (optional) A string to use as a prefix to GA Event Category.
	GaCategoryPrefix string

	// Should the Server log to stdout/err? Can be configured using IGN_LOGGER_LOG_STDOUT env var.
	LogToStd bool
	// Verbosity level of the Ign Logger - 4 debug, 3 info, 2 warning, 1 error, 0 critical
	LogVerbosity int
	// Verbosity level of the Ign Logger, to send to Rollbar - 4 debug, 3 info, 2 warning, 1 error, 0 critical
	RollbarLogVerbosity int
}

// IsUsingSSL returns true if the server was configured to use SSL.
func (s *Server) IsUsingSSL() bool {
	return s.SSLCert != "" && s.SSLKey != ""
}

// DatabaseConfig contains information about a database connection
type DatabaseConfig struct {
	// Username to login to a database.
	UserName string
	// Password to login to a database.
	Password string
	// Address of the database.
	Address string
	// Name of the database.
	Name string
	// Allowed Max Open Connections.
	// A value <= 0 means unlimited connections.
	// See 'https://golang.org/src/database/sql/sql.go'
	MaxOpenConns int
	// True to enable database logging. This will cause all database transactions
	// to output messages to standard out, which could create large log files in
	// docker. It is recommended to use this only during development and testing.
	// Logging can be controlled via the IGN_DB_LOG environment variable.
	//
	// By default logging is enabled only in tests with verbose flag.
	EnableLog bool
}

// gServer is an internal pointer to the Server.
var gServer *Server

// Init initialize this package.
// Note: This method does not configure the Server's Router. You will later
// need to configure the router and set it to the server.
func Init(auth0RSAPublicKey string, dbNameSuffix string, monitoring monitoring.Provider) (server *Server, err error) {

	// Configure and setup rollbar.
	// Do this first so that the logging connection is established.
	rollbarConfigure()

	server = &Server{
		HTTPPort:   ":8000",
		SSLport:    ":4430",
		monitoring: monitoring,
	}
	if err := server.readPropertiesFromEnvVars(); err != nil {
		return nil, err
	}

	gServer = server

	// Testing
	server.IsTest = strings.Contains(strings.ToLower(os.Args[0]), "test")

	if server.IsTest {
		// Let's use a separate DB name if under test mode.
		server.DbConfig.Name = server.DbConfig.Name + "_test"
		server.DbConfig.EnableLog = true
	}

	server.DbConfig.Name = server.DbConfig.Name + dbNameSuffix

	// Initialize the database
	err = server.dbInit()

	if err != nil {
		log.Println(err)
	}

	if server.IsTest {
		server.initTests()
	} else {
		server.SetAuth0RsaPublicKey(auth0RSAPublicKey)
	}

	return
}

// ConfigureRouterWithRoutes takes a given mux.Router and configures it with a set of
// declared routes. The router is configured with default middlewares.
// If a monitoring provider was set on the server, the router will include an additional middleware
// to track server metrics and add a monitoring route defined by the provider.
// If the router is a mux subrouter gotten by PathPrefix().Subrouter() then you need to
// pass the pathPrefix as argument here too (eg. "/2.0/")
func (s *Server) ConfigureRouterWithRoutes(pathPrefix string, router *mux.Router, routes Routes) {
	rc := NewRouterConfigurer(router, s.monitoring)
	rc.SetAuthHandlers(
		CreateJWTOptionalMiddleware(s),
		CreateJWTRequiredMiddleware(s),
	)
	rc.ConfigureRouter(pathPrefix, routes)
}

// SetRouter sets the main mux.Router to the server.
// If a monitoring provider has been defined, this will also configure
// the router to include routes for the monitoring service.
func (s *Server) SetRouter(router *mux.Router) *Server {
	if s.monitoring != nil {
		subrouter := router.PathPrefix("/").Subrouter()
		s.ConfigureRouterWithRoutes("/", subrouter, Routes{s.getMetricsRoute()})
	}

	s.Router = router
	return s
}

// ReadStdLogEnvVar reads the IGN_LOGGER_LOG_STDOUT env var and returns its bool value.
func ReadStdLogEnvVar() bool {
	// Get whether to enable logging to stdout/err
	strValue, err := ReadEnvVar("IGN_LOGGER_LOG_STDOUT")
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_LOG_STDOUT env variable. Disabling log to std.")
		return false
	}

	flag, err := strconv.ParseBool(strValue)
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_LOG_STDOUT env variable. Disabling log to std.")
		flag = false
	}
	return flag
}

// ReadLogVerbosityEnvVar reads the IGN_LOGGER_VERBOSITY env var and returns its bool value.
func ReadLogVerbosityEnvVar() int {
	// Get whether to enable logging to stdout/err
	strValue, err := ReadEnvVar("IGN_LOGGER_VERBOSITY")
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_VERBOSITY env variable. Using default values")
		return VerbosityWarning
	}

	val, err := strconv.ParseInt(strValue, 10, 32)
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_VERBOSITY env variable. Using default values")
		return VerbosityWarning
	}

	return int(val)
}

// ReadRollbarLogVerbosityEnvVar reads the IGN_LOGGER_ROLLBAR_VERBOSITY env var and returns its bool value.
func ReadRollbarLogVerbosityEnvVar() int {
	rollbarVerbStr, err := ReadEnvVar("IGN_LOGGER_ROLLBAR_VERBOSITY")
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_ROLLBAR_VERBOSITY env variable. Using WARNING as default value")
		return VerbosityWarning
	}

	val, err := strconv.ParseInt(rollbarVerbStr, 10, 32)
	if err != nil {
		log.Printf("Error parsing IGN_LOGGER_ROLLBAR_VERBOSITY env variable. Using WARNING as default value")
		return VerbosityWarning
	}
	return int(val)
}

// NewDatabaseConfigFromEnvVars returns a DatabaseConfig object from the following env vars:
// - IGN_DB_USERNAME
// - IGN_DB_PASSWORD
// - IGN_DB_ADDRESS
// - IGN_DB_NAME
// - IGN_DB_MAX_OPEN_CONNS - (Optional) You run the risk of getting a 'too many connections' error if this is not set.
func NewDatabaseConfigFromEnvVars() (DatabaseConfig, error) {
	dbConfig := DatabaseConfig{}
	var err error

	// Get the database username
	if dbConfig.UserName, err = ReadEnvVar("IGN_DB_USERNAME"); err != nil {
		errMsg := "Missing IGN_DB_USERNAME env variable. Database connection will not work"
		return dbConfig, errors.New(errMsg)
	}

	// Get the database password
	if dbConfig.Password, err = ReadEnvVar("IGN_DB_PASSWORD"); err != nil {
		errMsg := "Missing IGN_DB_PASSWORD env variable. Database connection will not work"
		return dbConfig, errors.New(errMsg)
	}

	// Get the database address
	if dbConfig.Address, err = ReadEnvVar("IGN_DB_ADDRESS"); err != nil {
		errMsg := "Missing IGN_DB_ADDRESS env variable. Database connection will not work"
		return dbConfig, errors.New(errMsg)
	}

	// Get the database name
	if dbConfig.Name, err = ReadEnvVar("IGN_DB_NAME"); err != nil {
		errMsg := "Missing IGN_DB_NAME env variable. Database connection will not work"
		return dbConfig, errors.New(errMsg)
	}

	// Get the database max open conns
	var maxStr string
	if maxStr, err = ReadEnvVar("IGN_DB_MAX_OPEN_CONNS"); err != nil {
		log.Printf("Missing IGN_DB_MAX_OPEN_CONNS env variable." +
			"Database max open connections will be set to unlimited," +
			"with the risk of getting 'too many connections' error.")
		dbConfig.MaxOpenConns = 0
	} else {
		var i int64
		i, err = strconv.ParseInt(maxStr, 10, 32)
		if err != nil || i <= 0 {
			log.Printf("Error parsing IGN_DB_MAX_OPEN_CONNS env variable." +
				"Database max open connections will be set to unlimited," +
				"with the risk of getting 'too many connections' error.")
			dbConfig.MaxOpenConns = 0
		} else {
			dbConfig.MaxOpenConns = int(i)
		}
	}

	return dbConfig, nil
}

// readPropertiesFromEnvVars configures the server based on env vars.
func (s *Server) readPropertiesFromEnvVars() error {
	var err error

	// Get the HTTP Port, if specified.
	if httpPort, err := ReadEnvVar("IGN_HTTP_PORT"); err != nil {
		log.Printf("Missing IGN_HTTP_PORT env variable. Server will use %s.\n", s.HTTPPort)
	} else {
		s.HTTPPort = httpPort
	}

	// Get the SSL Port, if specified.
	if sslPort, err := ReadEnvVar("IGN_SSL_PORT"); err != nil {
		log.Printf("Missing IGN_SSL_PORT env variable. Server will use %s.", s.SSLport)
	} else {
		s.SSLport = sslPort
	}

	// Get the SSL certificate, if specified.
	if s.SSLCert, err = ReadEnvVar("IGN_SSL_CERT"); err != nil {
		log.Printf("Missing IGN_SSL_CERT env variable. " +
			"Server will not be secure (no https).")
	}
	// Get the SSL private key, if specified.
	if s.SSLKey, err = ReadEnvVar("IGN_SSL_KEY"); err != nil {
		log.Printf("Missing IGN_SSL_KEY env variable. " +
			"Server will not be secure (no https).")
	}

	// Read Google Analytics parameters
	if s.GaTrackingID, err = ReadEnvVar("IGN_GA_TRACKING_ID"); err != nil {
		log.Print("Missing IGN_GA_TRACKING_ID env variable. GA will not be enabled")
	}
	if s.GaAppName, err = ReadEnvVar("IGN_GA_APP_NAME"); err != nil {
		log.Print("Missing IGN_GA_APP_NAME env variable. GA will not be enabled")
	}
	if s.GaCategoryPrefix, err = ReadEnvVar("IGN_GA_CAT_PREFIX"); err != nil {
		log.Print("Missing optional IGN_GA_CAT_PREFIX env variable.")
	}

	if s.DbConfig, err = NewDatabaseConfigFromEnvVars(); err != nil {
		log.Print(err.Error())
	}

	// Get whether to enable database logging
	var dbLogStr string
	s.DbConfig.EnableLog = false
	if dbLogStr, err = ReadEnvVar("IGN_DB_LOG"); err == nil {
		if s.DbConfig.EnableLog, err = strconv.ParseBool(dbLogStr); err != nil {
			log.Printf("Error parsing IGN_DB_LOG env variable." +
				"Database logging will be disabled.")
		}
	}

	// Get whether to enable logging to stdout/err and the verbosity level.
	s.LogToStd = ReadStdLogEnvVar()
	s.LogVerbosity = ReadLogVerbosityEnvVar()
	s.RollbarLogVerbosity = ReadRollbarLogVerbosityEnvVar()

	return nil
}

// Auth0RsaPublicKey return the Auth0 public key
func (s *Server) Auth0RsaPublicKey() string {
	return s.auth0RsaPublickey
}

// SetAuth0RsaPublicKey sets the server's Auth0 RSA public key
func (s *Server) SetAuth0RsaPublicKey(key string) {
	s.auth0RsaPublickey = key
	s.pemKeyString = "-----BEGIN CERTIFICATE-----\n" + s.auth0RsaPublickey +
		"\n-----END CERTIFICATE-----"
}

// Run the router and server
func (s *Server) Run() {

	if s.SSLCert != "" && s.SSLKey != "" {
		// Start the webserver with TLS support.
		log.Fatal(http.ListenAndServeTLS(s.SSLport, s.SSLCert, s.SSLKey, s.Router))
	} else {
		// Start the http webserver
		log.Fatal(http.ListenAndServe(s.HTTPPort, s.Router))
	}

	// Wait for all rollbar messages to complete
	rollbar.Wait()
}

// ///////////////////////////////////////////////
// Private functions

// initTests is run as the last step of init() and only when `go test` was run.
func (s *Server) initTests() {
	// Override Auth0 public RSA key with test key, if present
	if testKey, err := ReadEnvVar("TEST_RSA256_PUBLIC_KEY"); err != nil {
		log.Printf("Missing TEST_RSA256_PUBLIC_KEY. Test with authentication may not work.")
	} else {
		s.SetAuth0RsaPublicKey(testKey)
	}
}

// connectDb is called to establish a db connection. The target database is created if it doesn't exist.
func connectDb(driver, url string, cfg *DatabaseConfig) (db *gorm.DB, err error) {

	// Queries
	queryCreate := fmt.Sprintf("CREATE DATABASE %s", cfg.Name)
	queryUse := fmt.Sprintf("USE %s", cfg.Name)

	// Error message
	errUseDb := fmt.Sprintf("[ERROR] Unable to use the %s database", cfg.Name)

	// Open db connection
	db, err = gorm.Open(driver, url)

	if err != nil {
		return nil, errors.New("[ERROR] Unable to connect to the database system")
	}

	// Execute db creation
	_, err = db.DB().Exec(queryCreate)

	// If the step before does not throw any errors, it means that the database was successfully created
	// In the other hand, if it throws an error, it means that the database already exists
	if err == nil {
		log.Printf("[SUCCESS] Database %s was successfully created. Trying to connect once again...\n", cfg.Name)
	}

	// We have ensured that the database was created, let's use it.
	_, err = db.DB().Exec(queryUse)

	// If there was an error, it means that the database is not available
	if err != nil {
		return nil, errors.New(errUseDb)
	}

	// Close and reopen the DB to the correct database.
	db.Close()
	url = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=UTC",
		cfg.UserName, cfg.Password, cfg.Address, cfg.Name)
	db, _ = gorm.Open(driver, url)

	return db, err
}

// InitDbWithCfg initialize the database connection based on the given cfg.
func InitDbWithCfg(cfg *DatabaseConfig) (*gorm.DB, error) {
	// Connect to the database
	url := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8&parseTime=True&loc=UTC",
		cfg.UserName, cfg.Password, cfg.Address)

	var err error
	var db *gorm.DB

	for i := 0; i < 10; i++ {
		db, err = connectDb("mysql", url, cfg)

		if err == nil {
			break
		}

		log.Printf("Attempt[%d] to connect to the database failed.\n", i)
		log.Println(url)
		log.Println(err)
		time.Sleep(5 * time.Second)
	}

	return db, err
}

// dbInit Initialize the database connection
func (s *Server) dbInit() error {
	var err error
	s.Db, err = InitDbWithCfg(&s.DbConfig)
	// By default, assume this database is the User Database
	// Note: web-cloudsim uses a different default Db, and sets the UsersDb
	// appropriately in order to support access tokens.
	// \todo(anyone) Fix/remove this when the User database is moved to its own
	// server/service.
	s.UsersDb = s.Db
	return err
}

// rollbarConfigure setups up the rollbar connection.
func rollbarConfigure() {
	// token is the rollbar connection token.
	var token string

	// env is the environment string, usually "staging" or "production"
	var env string

	// Path to the application code root, not including the final slash.
	// Used to collapse non-project code when displaying tracebacks.
	var root string

	var err error

	// Get the rollbar token
	if token, err = ReadEnvVar("IGN_ROLLBAR_TOKEN"); err != nil {
		log.Printf("Missing IGN_ROLLBAR_TOKEN env variable." +
			"Rollbar connection will not work.")
		// Short circuit.
		return
	}

	// Get the rollbar environment
	if env, err = ReadEnvVar("IGN_ROLLBAR_ENV"); err != nil {
		log.Printf("Missing IGN_ROLLBAR_ENV env variable." +
			"Rollbar environment will be development.")
		env = "development"
	}

	// Get the rollbar root
	if root, err = ReadEnvVar("IGN_ROLLBAR_ROOT"); err != nil {
		log.Printf("Missing IGN_ROLLBAR_ROOT env variable." +
			"Rollbar will use bitbucket.org/ignitionrobotics.")
		root = "bitbucket.org/ignitionrobotics"
	}

	// Configure rollbar token.
	rollbar.SetToken(token)

	// Configure rollbar environment.
	rollbar.SetEnvironment(env)

	// Configure rollbar server root
	rollbar.SetServerRoot(root)

	// CodeVersion is a string, up to 40 characters, describing the version of
	// the application code. Rollbar understands these formats:
	// - semantic version (i.e. "2.1.12")
	// - integer (i.e. "45")
	// - SHA (i.e. "3da541559918a808c2402bba5012f6c60b27661c")
	//
	// We automatically acquire the version using mercurial
	rollbar.SetCodeVersion("unknown")
	_, filename, _, _ := runtime.Caller(1)
	if codeVersion, err := exec.Command("hg", "id", "-i", path.Dir(filename)).Output(); err == nil {
		rollbar.SetCodeVersion(string(codeVersion[:]))
	}

	// Set the server hostname
	if hostname, err := os.Hostname(); err == nil {
		rollbar.SetServerHost(hostname)
	} else {
		rollbar.SetServerHost("error reading hostname")
	}
}

// generateMetricsRoute is an internal method to generate a metrics route.
// This route is called by the monitoring system to scrape server metric data.
func (s *Server) getMetricsRoute() Route {
	return Route{
		Name:        "Metrics",
		Description: "Provides server metrics for monitoring systems.",
		URI:         s.monitoring.MetricsRoute(),
		Methods: Methods{
			Method{
				Type:        "GET",
				Description: "Get server metrics.",
				Handlers: FormatHandlers{
					FormatHandler{
						Extension: "",
						Handler:   s.monitoring.MetricsHandler(),
					},
				},
			},
		},
	}
}
