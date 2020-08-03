package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestWebsocketAddressSuite(t *testing.T) {
	suite.Run(t, &PrometheusProviderTestSuite{})
}

type PrometheusProviderTestSuite struct {
	suite.Suite
	provider  *provider
	testRoute string
}

func (suite *PrometheusProviderTestSuite) SetupSuite() {
	suite.provider = NewPrometheusProvider("").(*provider)
	suite.testRoute = "/test/route"
}

func (suite *PrometheusProviderTestSuite) TestNewPrometheusProvider() {
	route := ""
	p := NewPrometheusProvider(route).(*provider)
	suite.Equal(defaultMetricsRoute, p.route)

	route = suite.testRoute
	p = NewPrometheusProvider(route).(*provider)
	suite.Equal(route, p.route)
}

func (suite *PrometheusProviderTestSuite) TestMetricsRoute() {
	suite.Equal(defaultMetricsRoute, suite.provider.MetricsRoute())
}

func (suite *PrometheusProviderTestSuite) TestMetricsHandler() {
	expected := reflect.ValueOf(promhttp.Handler()).Pointer()
	result := reflect.ValueOf(suite.provider.MetricsHandler()).Pointer()
	suite.Equal(expected, result)
}

func (suite *PrometheusProviderTestSuite) TestMiddleware() {
	// Prometheus test metrics
	// RequestDurationSeconds
	requestDurationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Seconds spent serving HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	suite.provider.requestDurationSeconds = requestDurationSeconds

	// TotalRequests
	totalRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "The total number of HTTP requests.",
		},
		[]string{"status"},
	)
	suite.provider.totalRequests = totalRequests

	testcases := []struct {
		Method string
		Path   string
		Status string
		// CounterValue contains the expected value of the TotalRequests metric for the given labels after processing
		// each request
		CounterValue float64
	}{
		{
			Method:       "GET",
			Path:         suite.testRoute,
			Status:       "200",
			CounterValue: 1,
		},
		{
			Method:       "GET",
			Path:         suite.testRoute,
			Status:       "400",
			CounterValue: 1,
		},
		{
			Method:       "GET",
			Path:         suite.testRoute,
			Status:       "500",
			CounterValue: 1,
		},
		{
			Method:       "POST",
			Path:         suite.testRoute,
			Status:       "200",
			CounterValue: 2,
		},
		{
			Method:       "POST",
			Path:         suite.testRoute,
			Status:       "400",
			CounterValue: 2,
		},
		{
			Method:       "POST",
			Path:         suite.testRoute,
			Status:       "500",
			CounterValue: 2,
		},
	}

	middleware := suite.provider.Middleware()
	// Creates a mock HTTP handler
	makeHandler := func(status string) http.HandlerFunc {
		intStatus, err := strconv.Atoi(status)
		suite.NoError(err)

		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(intStatus)
		}
	}
	for _, tc := range testcases {
		// Check the values before sending the request
		counterValue := testutil.ToFloat64(suite.provider.totalRequests.WithLabelValues(tc.Status))
		suite.Equal(tc.CounterValue-1, counterValue)

		// Send requests
		req, err := http.NewRequest(tc.Method, "http://example.com"+tc.Path, nil)
		suite.NoError(err)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req, makeHandler(tc.Status))

		// Check that the counter has been updated
		counterValue = testutil.ToFloat64(suite.provider.totalRequests.WithLabelValues(tc.Status))
		suite.Equal(tc.CounterValue, counterValue)
	}
}
