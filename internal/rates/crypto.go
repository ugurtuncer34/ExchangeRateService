package rates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type BinanceResponse struct {
	Symbol string `json:"symbol"`
	Price string `json:"price"`
}

// object with TTL for crypto data
type CryptoCacheItem struct {
	Price float64
	ExpiresAt time.Time
}

// special thread-safe cache for crypto
type CryptoCache struct {
	mu sync.RWMutex
	items map[string]CryptoCacheItem
}

func NewCryptoCache() *CryptoCache {
	return &CryptoCache{
		items: make(map[string]CryptoCacheItem),
	}
}

// if in memory or TTL valid
func (c *CryptoCache) Get(symbol string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[symbol]
	if !exists {
		return 0, false // never fetched
	}

	// if time passed expiresAt, ignore data
	if time.Now().After(item.ExpiresAt) {
		return 0, false
	}

	return item.Price, true
}

// set price in memory and declare TTL
func (c *CryptoCache) Set(symbol string, price float64, ttlMinutes time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[symbol] = CryptoCacheItem{
		Price:     price,
		ExpiresAt: time.Now().Add(ttlMinutes * time.Minute),
	}
}

// fetch price from free Binance API
func FetchCryptoPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("binance api error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("binance api returned status: %d", resp.StatusCode)
	}

	var binanceResp BinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&binanceResp); err != nil {
		return 0, fmt.Errorf("failed to decode binance response: %v", err)
	}

	// string price to float
	price, err := strconv.ParseFloat(binanceResp.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse crypto price: %v", err)
	}

	return price, nil
}