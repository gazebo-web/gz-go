package gz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gazebo-web/gz-go/v10/monitoring"
	"github.com/gorilla/mux"
)

// NewRouter just creates a new Gorilla/mux router
func NewRouter() *mux.Router {
	// We need to set StrictSlash to "false" (default) to avoid getting
	// routes redirected automatically.
	return mux.NewRouter().StrictSlash(false)
}

// RouterConfigurer is used to configure a mux.Router with declared routes
// and middlewares. It also adds support for default global OPTIONS handler
// based on the route declarations.
type RouterConfigurer struct {
	// Embedded type mux.Router
	// See https://golang.org/doc/effective_go.html#embedding
	*mux.Router
	// An optional list of middlewares to be injected between the common
	// middlewares and the final route handler.
	customHandlers []negroni.Handler

	corsMap map[string]int
	// sortedREs keeps a sorted list of registered routes in corsMap.
	// It allows us to iterate the corsMap in 'order'.
	sortedREs []string
	// declared Routes
	routes *Routes
	// private field to keep a reference to JWT middlewares
	authOptMiddleware negroni.HandlerFunc
	authReqMiddleware negroni.HandlerFunc
	// monitoring provides a middleware to keep track of server request metrics.
	monitoring monitoring.Provider
}

// NewRouterConfigurer creates a new RouterConfigurer, used to
// configure a mux.Router with routes declarations.
func NewRouterConfigurer(r *mux.Router, monitoring monitoring.Provider) *RouterConfigurer {
	rc := &RouterConfigurer{
		Router:     r,
		corsMap:    make(map[string]int, 0),
		monitoring: monitoring,
	}
	return rc
}

// SetCustomHandlers - allows to set a list of optional middlewares
// that will be injected between the common middlewares and the final route handler.
func (rc *RouterConfigurer) SetCustomHandlers(handlers ...negroni.Handler) *RouterConfigurer {
	rc.customHandlers = handlers
	return rc
}

// SetAuthHandlers - set the JWT handlers to be used by the router for secure and unsecure
// routes.
func (rc *RouterConfigurer) SetAuthHandlers(optionalJWT, requiredJWT negroni.HandlerFunc) *RouterConfigurer {
	rc.authOptMiddleware = optionalJWT
	rc.authReqMiddleware = requiredJWT
	return rc
}

// ConfigureRouter - given an array of Route declarations,
// this method confifures the router to handle them.
// This is the main method to invoke with a RouterConfigurer.
//
// It supports an optional pathPrefix used to differentiate API versions (eg. "/2.0/").
func (rc *RouterConfigurer) ConfigureRouter(pathPrefix string, routes Routes) *RouterConfigurer {
	// Store the route declarations in the router.
	rc.routes = &routes

	for routeIndex, route := range routes {

		// Process unsecure routes
		for _, method := range route.Methods {
			for _, formatHandler := range method.Handlers {
				rc.registerRouteHandler(routeIndex, method.Type, false, formatHandler)
				rc.registerRouteInOptionsHandler(pathPrefix, routeIndex, formatHandler)
			}
		}

		// Process secure routes
		for _, method := range route.SecureMethods {
			for _, formatHandler := range method.Handlers {
				rc.registerRouteHandler(routeIndex, method.Type, true, formatHandler)
				rc.registerRouteInOptionsHandler(pathPrefix, routeIndex, formatHandler)
			}
		}
	}

	// Sorting corsMap is needed to correctly resolve OPTION requests
	// that need to match a regex.
	rc.sortedREs = getSortedREs(rc.corsMap)

	return rc
}

// ///////////////////////////////////////////////

// Internal method that registers the route (with its format)
// into the router's corsMap, for later use by the OPTIONS handler.
func (rc *RouterConfigurer) registerRouteInOptionsHandler(pathPrefix string,
	routeIndex int, formatHandler FormatHandler) {
	route := (*rc.routes)[routeIndex]
	// Setup a helper regex for "{_text_}" URL parameters.
	re := regexp.MustCompile("{[^}]+?}")
	namedVarRE := regexp.MustCompile("{[^}]+:[^{]+}")
	// Register the route in the corsMap. Used by the global OPTIONS handler
	uriPath := route.URI + formatHandler.Extension
	prefixedURIPath := strings.Replace(pathPrefix+uriPath, "//", "/", -1)
	// Store route information for the global OPTIONS handler
	newStr := strings.Replace(prefixedURIPath, ".", "\\.", -1)
	reString := namedVarRE.ReplaceAllString(newStr, ".+")
	reString = re.ReplaceAllString(reString, "[^/]+")
	rc.corsMap[reString] = routeIndex

	rc.
		Methods("OPTIONS").
		Path(uriPath).
		Name(route.Name + formatHandler.Extension).
		Handler(http.HandlerFunc(rc.globalOptionsHandler))
}

// Helper function that registers the given route handler AND
// automatically creates and registers an HTTP OPTIONS method handler on the route.
//
// formatHandler is the given most route handler.
func (rc *RouterConfigurer) registerRouteHandler(routeIndex int, methodType string,
	secure bool, formatHandler FormatHandler) {

	handler := formatHandler.Handler

	// TODO move to top chain middlewares

	// Configure auth middleware
	var authMiddleware negroni.HandlerFunc
	if !secure {
		authMiddleware = rc.authOptMiddleware
	} else {
		authMiddleware = rc.authReqMiddleware
	}

	routeName := (*rc.routes)[routeIndex].Name

	recovery := negroni.NewRecovery()
	// PrintStack is set to false to avoid sending stacktrace to client.
	recovery.PrintStack = false

	// Configure middleware chain
	middlewares := negroni.New()
	// Add monitoring middleware if monitoring provider is present.
	// It must be the first middleware in the chain
	if rc.monitoring != nil {
		middlewares = middlewares.With(rc.monitoring.Middleware())
	}
	// Add default middlewares
	middlewares = middlewares.With(
		recovery,
		negroni.HandlerFunc(requireDBMiddleware),
		negroni.HandlerFunc(addCORSheadersMiddleware),
		authMiddleware,
		negroni.HandlerFunc(newGaEventTracking(routeName)),
	)
	// Inject custom handlers just before the final handler
	middlewares = middlewares.With(rc.customHandlers...)
	middlewares.Use(negroni.Wrap(http.Handler(handler)))

	// Last, wrap everything with a Logger middleware
	handler = logger(middlewares, routeName)

	uriPath := (*rc.routes)[routeIndex].URI + formatHandler.Extension

	// Create the route handler.
	rc.
		Methods(methodType).
		Path(uriPath).
		Name(routeName + formatHandler.Extension).
		Handler(handler)
}

// globalOptionsHandler is an OPTIONS method handler that will return
// documentation of a route based on its Route definition.
func (rc *RouterConfigurer) globalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	index := 0
	ok := false
	// Find the matching URL
	for _, key := range rc.sortedREs {
		// Make sure the regular expression matches the complete URL path
		if regexp.MustCompile(key).FindString(r.URL.Path) == r.URL.Path {
			ok = true
			index = rc.corsMap[key]
			break
		}
	}
	route := (*rc.routes)[index]
	if ok {
		if output, e := json.Marshal(route); e != nil {
			err := NewErrorMessageWithBase(ErrorMarshalJSON, e)
			reportJSONError(w, r, *err)
		} else {
			// Find all the route supported HTTP Methods
			allow := make([]string, 0)
			for _, m := range route.Methods {
				allow = append(allow, m.Type)
			}
			for _, m := range route.SecureMethods {
				allow = append(allow, m.Type)
			}
			w.Header().Set("Allow", strings.Join(allow[:], ","))
			w.Header().Set("Content-Type", "application/json")
			addCORSheaders(w)
			fmt.Fprintln(w, string(output))
		}
		return
	}

	// Return error if a URL did not match
	err := ErrorMessage(ErrorNameNotFound)
	reportJSONError(w, r, err)
}

// ///////////////////////////////////////////////
// sortRE is an internal []string wrapper type used to sort by
// the number of "[^/]+" string occurrences found in a regex (ie. count).
// If the same count is found then the larger string will take precedence.
type sortRE []string

func (s sortRE) Len() int {
	return len(s)
}

func (s sortRE) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortRE) Less(i, j int) bool {
	ci := strings.Count(s[i], "[^/]+")
	cj := strings.Count(s[j], "[^/]+")
	if ci == cj {
		return len(s[i]) > len(s[j])
	}
	return ci < cj
}

func getSortedREs(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(sortRE(keys))
	return keys
}
