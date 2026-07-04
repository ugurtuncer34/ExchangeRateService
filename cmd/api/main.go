package main

import (
	"exchangerateservice/internal/rates"
	"fmt"
	"log"
)

func main(){
	fmt.Println("Exchange Rate Service (Go) is up and ready!")

	fmt.Println("Latest EUR rate is coming from TCMB...")

	rate, err := rates.FetchTodayRate("EUR")
	if err != nil {
		log.Fatalf("An error occured: %v", err) // log the error on screen and exit 1
	}

	fmt.Printf("Success! Latest EUR rate: %f TL\n", rate)
}