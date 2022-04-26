package net

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// DefaultMiddlewares groups a set of middlewares used across different services.
var DefaultMiddlewares = []Middleware{
	middleware.RequestID,
	middleware.RealIP,
	// TODO: Uncomment this line when solving issue with health-checks.
	//middleware.Logger,
	middleware.Recoverer,
}

// Middleware is function that gets executed before each HTTP handler.
// It's useful for adding request-specific logic to every request like logging and setting request id.
type Middleware func(handler http.Handler) http.Handler

// NewRouter initializes a new HTTP router using chi.
// It also loads middlewares used on every incoming HTTP request.
func NewRouter(middlewares ...Middleware) chi.Router {
	r := chi.NewRouter()
	for _, m := range middlewares {
		r.Use(m)
	}
	return r
}

// RoutesGetter holds a method to get routes.
type RoutesGetter interface {
	// Routes returns a set of routes in a http.Handler form.
	Routes() http.Handler
}
