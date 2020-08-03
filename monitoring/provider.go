package monitoring

import (
	"github.com/urfave/negroni"
	"net/http"
)

// Provider is a piece of middleware that acts as a provider for a monitoring system.
// This includes setting up setting up a middleware to gather data and an endpoint on the server that the monitoring
// system can use to scrape data.
type Provider interface {
	// MetricsRoute returns the route to the metrics endpoint.
	MetricsRoute() string
	// MetricsHandler returns an HTTP handler to use with the metrics endpoint.
	MetricsHandler() http.Handler
	// Middleware returns a middleware used to gather metrics from HTTP traffic.
	Middleware(args ...interface{}) negroni.Handler
}
