package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/leogsouza/api-remote-job-salary-calculator/logger"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("accepting connections on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), handler()); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

func handler() http.Handler {
	limiter := tollbooth.NewLimiter(1, nil)
	r := chi.NewRouter()

	logrus := logger.New()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.NewStructuredLogger(logrus))
	r.Use(middleware.Recoverer)
	r.Use(tollbooth_chi.LimitHandler(limiter))
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Heartbeat("/"))

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Get("/salary/calculator", calculateHandler)

	return r
}
