package rates

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type RateResponse struct { // response json struct
	Currency string `json:"currency"`
	Date string `json:"date"`
	Rate float64 `json:"rate"`
	Source string `json:"source"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// this struct moves cache (dependency) to handler func. like dependency injection in .Net
type RateHandler struct {
	Cache *RateCache
}

// HTTP request greeting func
func (h *RateHandler) GetRate(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	dateStr := r.URL.Query().Get("date")

	if currency == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'currency' parameter (e.g. EUR)")
		return
	}
	if dateStr == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'date' parameter (format: YYYY-MM-DD)")
		return
	}

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// cache control
	cacheKey := fmt.Sprintf("%s_%s", currency, targetDate.Format("2006-01-02"))

	if cachedRate, exists := h.Cache.Get(cacheKey); exists {
		// found in memory
		response := RateResponse{
			Currency: currency,
			Date: dateStr,
			Rate: cachedRate,
			Source: "Cache", // just to see
		}
		log.Printf("Served %s from CACHE", cacheKey)
		writeJSON(w, http.StatusOK, response)
		return
	}
	
	// if not in cache, fetch from TCMB
	rate, err := FetchRateByDate(currency, targetDate)
	if err != nil {
		log.Printf("TCMB fetch error: %v", err)
		writeJSONError(w, http.StatusBadGateway, fmt.Sprintf("Failed to fetch rate: %v", err))
		return
	}

	// write to cache
	h.Cache.Set(cacheKey, rate)

	// return success json
	response := RateResponse{
		Currency: currency,
		Date: dateStr,
		Rate: rate,
		Source: "TCMB",
	}

	log.Printf("Served %s from TCMB", cacheKey)
	writeJSON(w, http.StatusOK, response)
}

// helper func: returns success json
func writeJSON(w http.ResponseWriter, status int, data interface{}){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// helper error json func
func writeJSONError(w http.ResponseWriter, status int, message string){
	writeJSON(w, status, ErrorResponse{Error: message})
}