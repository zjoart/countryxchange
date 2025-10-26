package countries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/zjoart/countryxchange/pkg/logger"
)

const (
	countriesURL = "https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies"
	ratesURL     = "https://open.er-api.com/v6/latest/USD"
)

// RefreshResult summarizes a refresh operation
type RefreshResult struct {
	Total         int
	LastRefreshed time.Time
}

// external structs
type restCountry struct {
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Region     string `json:"region"`
	Population int64  `json:"population"`
	Flag       string `json:"flag"`
	Currencies []struct {
		Code string `json:"code"`
	} `json:"currencies"`
}

type ratesResp struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// ExternalError marks which external API failed
type ExternalError struct {
	API string
}

func (e ExternalError) Error() string {
	return fmt.Sprintf("Could not fetch data from %s", e.API)
}

// Refresh fetches external data and updates DB in a transaction.
// If external fetch fails, no DB changes are made.
func Refresh(ctx context.Context, db *sql.DB) (*RefreshResult, error) {
	logger.Info("service: Refresh started")
	client := &http.Client{Timeout: 20 * time.Second}

	// fetch countries
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, countriesURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("service: failed fetching countries", logger.WithError(err))
		return nil, ExternalError{API: "restcountries"}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("service: restcountries returned non-200", logger.Fields{"status": resp.StatusCode})
		return nil, ExternalError{API: "restcountries"}
	}

	var rc []restCountry
	if err := json.NewDecoder(resp.Body).Decode(&rc); err != nil {
		logger.Warn("service: failed decoding restcountries response", logger.WithError(err))
		return nil, ExternalError{API: "restcountries"}
	}

	// fetch rates
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, ratesURL, nil)
	resp2, err := client.Do(req2)
	if err != nil {
		logger.Warn("service: failed fetching exchange rates", logger.WithError(err))
		return nil, ExternalError{API: "exchangerates"}
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		logger.Warn("service: exchangerates returned non-200", logger.Fields{"status": resp2.StatusCode})
		return nil, ExternalError{API: "exchangerates"}
	}

	var rr ratesResp
	if err := json.NewDecoder(resp2.Body).Decode(&rr); err != nil {
		logger.Warn("service: failed decoding exchangerates response", logger.WithError(err))
		return nil, ExternalError{API: "exchangerates"}
	}

	// prepare DB
	if err := EnsureTables(db); err != nil {
		logger.Error("service: EnsureTables failed", logger.WithError(err))
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("service: begin tx failed", logger.WithError(err))
		return nil, err
	}

	// seed rand
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	now := time.Now().UTC()

	processed := 0
	for _, rcountry := range rc {
		// prepare Country struct for validation
		if rcountry.Name == "" {
			logger.Warn("service: country name missing from external API")
			continue
		}

		var currencyCode *string
		var exchangeRate *float64
		var estimatedGDP *float64

		if len(rcountry.Currencies) > 0 && rcountry.Currencies[0].Code != "" {
			code := rcountry.Currencies[0].Code
			currencyCode = &code
			if rate, ok := rr.Rates[code]; ok {
				exchangeRate = &rate
				// compute estimated_gdp = population * random(1000-2000) / exchange_rate
				mult := float64(r.Intn(1001) + 1000) // 1000..2000
				est := float64(rcountry.Population) * mult / (*exchangeRate)
				estimatedGDP = &est
			} else {
				// not found in rates => leave exchangeRate nil and estimated_gdp nil
				exchangeRate = nil
				estimatedGDP = nil
			}
		} else {
			// currencies empty
			currencyCode = nil
			exchangeRate = nil
			zero := 0.0
			estimatedGDP = &zero
		}

		c := &Country{
			Name:            rcountry.Name,
			Population:      rcountry.Population,
			LastRefreshedAt: &now,
		}
		if rcountry.Capital != "" {
			c.Capital = &rcountry.Capital
		}
		if rcountry.Region != "" {
			c.Region = &rcountry.Region
		}
		if rcountry.Flag != "" {
			c.FlagURL = &rcountry.Flag
		}
		c.CurrencyCode = currencyCode
		c.ExchangeRate = exchangeRate
		c.EstimatedGDP = estimatedGDP

		// validate before upserting
		if err := c.Validate(); err != nil {
			logger.Warn("service: country validation failed", logger.Fields{
				"country": c.Name,
				"errors":  err.(*ValidationError).Errors,
			})
			continue
		}

		if err := UpsertCountry(tx, c); err != nil {
			logger.Error("service: UpsertCountry failed", logger.WithError(err))
			tx.Rollback()
			return nil, err
		}
		processed++
	}

	// save last refreshed
	if err := SaveLastRefreshed(tx, now); err != nil {
		logger.Error("service: SaveLastRefreshed failed", logger.WithError(err))
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("service: tx commit failed", logger.WithError(err))
		tx.Rollback()
		return nil, err
	}

	// generate summary image (best-effort)
	go func() {
		if err := GenerateSummaryImage(db, "cache/summary.png"); err != nil {
			logger.Warn("service: GenerateSummaryImage failed", logger.WithError(err))
		} else {
			logger.Info("service: GenerateSummaryImage completed")
		}
	}()

	logger.Info("service: Refresh completed", logger.Fields{"total_processed": processed})
	return &RefreshResult{Total: processed, LastRefreshed: now}, nil
}
