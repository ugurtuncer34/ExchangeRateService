package rates

import (
	"fmt"
	"log"
	"time"
)

// Cron job to fetch USD and EUR every work day at 3:30pm
func StartProactiveCache(cache *RateCache) {
	// TCMB works on UTC+3
	trtZone := time.FixedZone("TRT", 3*60*60)

	go func() { // Goroutine, separated from main flow
		for {
			now := time.Now().In(trtZone)

			target := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, trtZone)

			// if time passed 15:30 or today is weekend, make target one day forward
			if now.After(target) || now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
				target = target.Add(24 * time.Hour)
			}

			if target.Weekday() == time.Saturday {
				target = target.Add(48 * time.Hour)
			} else if target.Weekday() == time.Sunday {
				target = target.Add(24 * time.Hour)
			}

			// calculate sleep duration
			duration := target.Sub(now)
			log.Printf("Background worker sleeping for %v until next scheduled fetch at %s", duration, target.Format("2006-01-02 15:04:05"))

			time.Sleep(duration) // CPU consuming down to 0%

			log.Println("Background worker woke up! Fetching EUR and USD...")

			currencies := []string{"EUR", "USD"}
			fetchDate := time.Now().In(trtZone)

			for _, curr := range currencies {
				rate, err := FetchRateByDate(curr, fetchDate)
				if err != nil {
					log.Printf("Background fetch failed for %s: %v", curr, err)
					continue
				}

				// successfully retrieved, write to cache
				cacheKey := fmt.Sprintf("%s_%s", curr, fetchDate.Format("2006-01-02"))
				cache.Set(cacheKey, rate)
				log.Printf("Proactively cached %s: %f", cacheKey, rate)
			}
		}
	}()
}