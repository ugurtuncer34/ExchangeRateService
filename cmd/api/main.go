package main

import (
	"exchangerateservice/internal/rates"
	"log"
	"time"
)

func main(){
	log.Println("Exchange Rate Service (Go) is starting...")

	// YYYY-MM-DD, 2026-06-13 is saturday
	targetDateString := "2026-06-13"

	// convert string date to time.Time object in Go
	targetDate, err := time.Parse("2006-01-02", targetDateString)
	if err != nil {
		log.Fatalf("Failed to parse date: %v", err)
	}

	currencyCode := "EUR"
	log.Printf("Requesting %s rate for date: %s", currencyCode, targetDateString)

	rate, err := rates.FetchTodayRate(currencyCode, targetDate)
	if err != nil {
		log.Fatalf("Error occured: %v", err)
	}

	log.Printf("Success! Retrieved %s rate: %f", currencyCode, rate)
}