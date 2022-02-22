package net

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"net/http"
)

// NewRouter initializes a new HTTP router using chi.
// It also loads middlewares used on every incoming HTTP request.
func NewRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	return r
}

// RoutesGetter holds a method to get routes.
type RoutesGetter interface {
	// Routes returns a set of routes in a http.Handler form.
	Routes() http.Handler
}
