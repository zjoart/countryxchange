package countries

import "time"

// Country represents a country record stored in the DB and returned by the API
type Country struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Capital         *string    `json:"capital,omitempty"`
	Region          *string    `json:"region,omitempty"`
	Population      int64      `json:"population"`
	CurrencyCode    *string    `json:"currency_code,omitempty"`
	ExchangeRate    *float64   `json:"exchange_rate,omitempty"`
	EstimatedGDP    *float64   `json:"estimated_gdp,omitempty"`
	FlagURL         *string    `json:"flag_url,omitempty"`
	LastRefreshedAt *time.Time `json:"last_refreshed_at,omitempty"`
}
