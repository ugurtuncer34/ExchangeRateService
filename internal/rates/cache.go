package rates

import "sync"

type RateCache struct { // if name starts with capital letter, it is public for packages
	mu sync.RWMutex
	rates map[string]float64
}

// factory (manufacture func). create empty map on memory and return struct pointer
func NewRateCache() *RateCache {
	return &RateCache{
		rates: make(map[string]float64),
	}
}

// Get returns the rate for the given key, bool indicates if the key is present in map
func (c *RateCache) Get(key string) (float64, bool) {
	c.mu.RLock() // start read lock (doesn't prevent other readers)
	defer c.mu.RUnlock() // unlock when func ends

	val, exists := c.rates[key]
	return val, exists
}

// Set adds or updates a rate
func (c *RateCache) Set(key string, value float64) {
	c.mu.Lock() // no one can read until write ends
	defer c.mu.Unlock()

	c.rates[key] = value
}