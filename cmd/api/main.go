package main

import (
	"exchangerateservice/internal/rates"
	"fmt"
	"log"
)

func main(){
	fmt.Println("Exchange Rate Service (Go) is starting...")

	cache := rates.NewRateCache()

	currencyCode := "EUR"

	// first request (cache is empty, goes to TCMB)
	log.Printf("1. Request: Fetching %s rate...", currencyCode)
	rate, exists := cache.Get(currencyCode) // first, ask cache
	if !exists {
		log.Println("Cache miss! Fetching from TCMB...")

		var err error
		rate, err = rates.FetchTodayRate(currencyCode)
		if err != nil {
			log.Fatalf("Error occured: %v", err)
		}

		// write value to cache
		cache.Set(currencyCode, rate)
		log.Printf("Successfully saved to cache. Rate: %f", rate)
	}

	// second request (same currency)
	log.Printf("2. Request: Fetching %s rate again...", currencyCode)
	rate2, exists2 := cache.Get(currencyCode)
	if exists2 {
		log.Printf("Cache hit! Retrieved from memory instantly. Rate: %f", rate2)
	} else {
		log.Println("This should not happen, it must be in cache!")
	} // used log instead of fmt because log.Println adds date and time by default
}