package rates

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// if date is weekend, go back to friday. (lower case letter, private func, available only in this file)
func getValidWorkday(date time.Time) time.Time {
	if date.Weekday() == time.Saturday {
		return date.AddDate(0, 0, -1)
	}
	if date.Weekday() == time.Sunday {
		return date.AddDate(0, 0, -2)
	}
	return date
}

// create TCMB url according to target date
func buildUrl(targetDate time.Time) string {
	now := time.Now()

	// if today, use today.xml
	if targetDate.Year() == now.Year() && targetDate.Month() == now.Month() && targetDate.Day() == now.Day() {
		return "https://www.tcmb.gov.tr/kurlar/today.xml"
	}
	// if not, generate appropriate directories for TCMB
	yearMonth := targetDate.Format("200601") // YYYYMM
	dayMonthYear := targetDate.Format("02012006") // DDMMYYYY

	return fmt.Sprintf("https://www.tcmb.gov.tr/kurlar/%s/%s.xml", yearMonth, dayMonthYear)
}

func FetchRateByDate(targetCurrency string, date time.Time) (float64, error) {
	validDate := getValidWorkday(date)
	url := buildUrl(validDate)

	resp, err := http.Get(url) // go can have multiple returns
	if err != nil { // go doesn't have trycatch
		return 0, fmt.Errorf("Couldn't connect to TCMB: %v", err)
	}
	defer resp.Body.Close() // do this immediately after this func finishes

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("TCMB returned bad status code %d for url: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("Couldn't read XML: %v", err)
	}

	var tcmbData TcmbResponse // Unmarshal read data to struct
	err = xml.Unmarshal(body, &tcmbData) // & is pointer like in C
	if err != nil {
		return 0, fmt.Errorf("Couldn't parse XML: %v", err)
	}

	// find the currency with for loop. "_" is index
	for _, c := range tcmbData.Currencies {
		if c.Code == targetCurrency {
			rate, err := strconv.ParseFloat(c.ForexBuying, 64) // parse string to float64
			if err != nil {
				return 0, fmt.Errorf("Currency value not numeric: %v", err)
			}
			return rate, nil // success, rate value and zero error
		}
	}
	// if not found after loop
	return 0, fmt.Errorf("Requested currency not found: %s", targetCurrency)
}