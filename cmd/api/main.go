package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"exchangerateservice/internal/rates"
)

func main(){
	r := chi.NewRouter()

	// useful middlewares (logging and crash preventing)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// wake up cache system
	cache := rates.NewRateCache()

	// start background worker
	rates.StartProactiveCache(cache)

	// create handler and inject cache dependency
	rateHandler := rates.RateHandler{
		Cache: cache,
	}

	// define endpoint
	r.Get("/api/rates", rateHandler.GetRate)

	// start server
	port := ":8080"
	log.Printf("Exchange Rate Service API is running on port %s", port)

	// http.ListenAndServe blocks the code, listens until close
	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}