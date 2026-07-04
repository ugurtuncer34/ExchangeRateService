package rates

import "encoding/xml"

type TcmbResponse struct {
	XMLName xml.Name `xml:"Tarih_Date"`
	Currencies []Currency `xml:"Currency"`
}

type Currency struct {
	Code string `xml:"Kod,attr"`
	ForexBuying string `xml:"ForexBuying"`
}