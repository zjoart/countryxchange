package countries

import "time"

// ValidationError represents field-level validation errors
type ValidationError struct {
	Errors map[string]string `json:"details"`
}

func (v *ValidationError) Error() string {
	return "Validation failed"
}

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

// Validate ensures required fields are present and valid
func (c *Country) Validate() error {
	errors := make(map[string]string)

	if c.Name == "" {
		errors["name"] = "is required"
	}
	if c.Population <= 0 {
		errors["population"] = "must be positive"
	}
	if c.CurrencyCode == nil || *c.CurrencyCode == "" {
		errors["currency_code"] = "is required"
	}

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}
	return nil
}
