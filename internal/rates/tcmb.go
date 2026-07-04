package rates

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func FetchTodayRate(targetCurrency string) (float64, error) {
	url := "https://www.tcmb.gov.tr/kurlar/today.xml"
	resp, err := http.Get(url) // go can have multiple returns

	if err != nil { // go doesn't have trycatch
		return 0, fmt.Errorf("Couldn't connect to TCMB: %v", err)
	}

	defer resp.Body.Close() // do this immediately after this func finishes

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Error code from TCMB: %d", resp.StatusCode)
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