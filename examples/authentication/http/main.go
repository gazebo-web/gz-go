package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/gazebo-web/gz-go/v9/middleware"
	"github.com/rs/cors"
)

func main() {
	// Get the public key location
	fp := os.Getenv("AUTHENTICATION_PUBLIC_KEY_FILE")

	// Read the content of the public key
	f, err := os.ReadFile(fp)
	if err != nil {
		log.Fatalln("Failed to read file:", err)
	}

	// Initialize Authentication component
	auth := authentication.NewAuth0(f)

	// Set up bearer token middleware
	bearer := middleware.BearerToken(auth)

	// Define HTTP handler
	h := http.Handler(http.HandlerFunc(handler))

	// Wrap handler with bearer token middleware
	h = bearer(h)

	// Set up CORS policy to Allow-All.
	h = cors.AllowAll().Handler(h)

	// Set up HTTP server and assign the respective HTTP handler
	srv := http.Server{
		Addr:         ":3030",
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	// Listen and serve on port 3030
	log.Println("Listening on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln("Failed to listen and serve:", err)
	}
	log.Println("Shutting server down...")
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Welcome!")); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
