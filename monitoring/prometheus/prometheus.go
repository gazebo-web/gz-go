package prometheus

import (
	"github.com/gazebo-web/gz-go/v8/monitoring"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/negroni"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultMetricsRoute is the default route used for the metrics handler Prometheus calls.
	defaultMetricsRoute = "/metrics"
)

var (
	// RequestDurationSeconds tracks the seconds spent serving HTTP requests.
	// Used to monitor service status and raise alerts (using the RED method)
	RequestDurationSeconds *prometheus.HistogramVec
	// TotalRequests tracks the total number of HTTP requests.
	// Used for auto-scaling
	TotalRequests *prometheus.CounterVec

	// invalidCharsRE is a regex used to convert a route to a compatible Prometheus label value
	invalidCharsRE = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// provider is an implementation of a monitoring provider that generates Prometheus metrics.
type provider struct {
	// route is the route that the metrics handle will be served from.
	route string
	// requestDurationSeconds contains the RequestDurationSeconds metric
	requestDurationSeconds *prometheus.HistogramVec
	// totalRequests contains the TotalRequests metric
	totalRequests *prometheus.CounterVec
}

// init initializes Prometheus metrics.
func init() {
	// RequestDurationSeconds
	RequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Seconds spent serving HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	prometheus.MustRegister(RequestDurationSeconds)

	// TotalRequests
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "The total number of HTTP requests.",
		},
		[]string{"status"},
	)
	prometheus.MustRegister(TotalRequests)
}

// NewPrometheusProvider creates a new Prometheus metrics provider.
// route is the route the Prometheus server will contact to get metric information.
// If route is an empty string, it will default to "/metrics".
// TODO Provide a mechanism to define additional application-specific metrics.
func NewPrometheusProvider(route string) monitoring.Provider {
	// Default route
	if route == "" {
		route = defaultMetricsRoute
	}

	return &provider{
		route:                  route,
		requestDurationSeconds: RequestDurationSeconds,
		totalRequests:          TotalRequests,
	}
}

// MetricsRoute returns the route to the metrics endpoint.
func (p *provider) MetricsRoute() string {
	return p.route
}

// MetricsHandler returns an HTTP handler to use with the metrics endpoint.
// It uses the default Prometheus handler.
func (p *provider) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// Middleware returns a middleware that is used to gather metrics from incoming requests.
func (p *provider) Middleware(args ...interface{}) negroni.Handler {
	return negroni.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			begin := time.Now()

			// An interceptor is used to access HTTP response properties
			interceptor := negroni.NewResponseWriter(w)

			// Process the request
			next.ServeHTTP(interceptor, r)

			// Process request and response values
			path := p.getRouteName(r)
			status := strconv.Itoa(interceptor.Status())
			took := time.Since(begin).Seconds()

			// Register metric data
			p.requestDurationSeconds.WithLabelValues(r.Method, path, status).Observe(took)
			p.totalRequests.WithLabelValues(status).Inc()
		},
	)
}

// getRouteName converts routes from a path to a string compatible with a Prometheus label value.
// Example: '/api/delay/{example}' to 'api_delay_example'
func (p *provider) getRouteName(r *http.Request) string {
	if mux.CurrentRoute(r) != nil {
		if name := mux.CurrentRoute(r).GetName(); len(name) > 0 {
			return p.urlToLabel(name)
		} else if path, err := mux.CurrentRoute(r).GetPathTemplate(); err == nil && len(path) > 0 {
			return p.urlToLabel(path)
		}
	}
	return p.urlToLabel(r.URL.Path)
}

// urlToLabel converts a URL path to a string compatible with a Prometheus label value.
func (p *provider) urlToLabel(path string) string {
	result := invalidCharsRE.ReplaceAllString(path, "_")
	result = strings.ToLower(strings.Trim(result, "_"))
	if result == "" {
		result = "root"
	}
	return result
}
