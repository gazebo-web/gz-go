package gz

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gazebo-web/gz-go/v10/monitoring"
	"github.com/gazebo-web/gz-go/v10/monitoring/prometheus"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestMonitoringPrometheusTestSuite(t *testing.T) {
	suite.Run(t, &MonitoringPrometheusTestSuite{})
}

type MonitoringPrometheusTestSuite struct {
	suite.Suite
	monitoring   monitoring.Provider
	metricsRoute string
	router       *mux.Router
	server       *Server
}

func (suite *MonitoringPrometheusTestSuite) SetupSuite() {
	suite.monitoring = prometheus.NewPrometheusProvider("")
	suite.metricsRoute = suite.monitoring.MetricsRoute()

	suite.server = &Server{
		HTTPPort:   ":8000",
		SSLport:    ":4430",
		monitoring: suite.monitoring,
	}

	// Prepare the router
	suite.router = NewRouter()
	// Test route
	suite.server.ConfigureRouterWithRoutes("/", suite.router, Routes{
		{
			Name:        "test",
			Description: "Test route.",
			URI:         "/test",
			Headers:     nil,
			Methods: Methods{
				{
					Type:        "GET",
					Description: "Test GET route.",
					Handlers: FormatHandlers{
						{
							Extension: "",
							Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								// Set the body
								_, err := w.Write([]byte("OK"))

								// Set the status code
								code := 200
								if err != nil {
									code = 500
								}
								w.WriteHeader(code)
							}),
						},
					},
				},
			},
		},
	})
	// SetRouter automatically configures the router with a monitoring route
	suite.server.SetRouter(suite.router)

	// Server needs to be initialized to prevent the logging middleware from panicking
	gServer = suite.server
	suite.server.Db = &gorm.DB{}
}

func (suite *MonitoringPrometheusTestSuite) TestMonitoringRouteExists() {
	// Check that the route exists and returns successfully
	match := &mux.RouteMatch{}
	req, err := http.NewRequest("GET", suite.metricsRoute, nil)
	suite.NoError(err)

	suite.True(suite.router.Match(req, match))
	suite.NoError(match.MatchErr)
}

func (suite *MonitoringPrometheusTestSuite) TestMonitoringRoute() {
	sendRequest := func(route string) *httptest.ResponseRecorder {
		req, err := http.NewRequest("GET", route, nil)
		suite.NoError(err)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		suite.Equal(200, w.Code)

		return w
	}
	responseBody := sendRequest("/test").Body.String()
	suite.Equal("OK", responseBody)
	responseBody = sendRequest("/test").Body.String()
	suite.Equal("OK", responseBody)

	responseBody = sendRequest("/metrics").Body.String()
	suite.Contains(responseBody, "http_requests_total{status=\"200\"} 2")
	suite.Contains(responseBody, "http_request_duration_seconds_count{method=\"GET\",path=\"test\",status=\"200\"} 2")

}
