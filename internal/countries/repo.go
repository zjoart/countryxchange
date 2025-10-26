package countries

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

// EnsureTables creates countries and metadata tables when needed
func EnsureTables(db *sql.DB) error {
	// countries table
	createCountries := `
    CREATE TABLE IF NOT EXISTS countries (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        capital VARCHAR(255),
        region VARCHAR(255),
        population BIGINT NOT NULL,
        currency_code VARCHAR(32),
        exchange_rate DOUBLE,
        estimated_gdp DOUBLE,
        flag_url VARCHAR(512),
        last_refreshed_at DATETIME,
        UNIQUE KEY unique_name (name)
    );`

	if _, err := db.Exec(createCountries); err != nil {
		return err
	}

	// metadata table for storing global values like last refresh
	createMeta := `
    CREATE TABLE IF NOT EXISTS metadata (
        meta_key VARCHAR(128) PRIMARY KEY,
        meta_value VARCHAR(1024),
        updated_at DATETIME
    );`

	if _, err := db.Exec(createMeta); err != nil {
		return err
	}

	return nil
}

// UpsertCountry inserts or updates country by name (unique)
func UpsertCountry(tx *sql.Tx, c *Country) error {
	q := `INSERT INTO countries
        (name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            capital = VALUES(capital),
            region = VALUES(region),
            population = VALUES(population),
            currency_code = VALUES(currency_code),
            exchange_rate = VALUES(exchange_rate),
            estimated_gdp = VALUES(estimated_gdp),
            flag_url = VALUES(flag_url),
            last_refreshed_at = VALUES(last_refreshed_at)
    `

	var capital, region, currency, flag sql.NullString
	var exchange, est sql.NullFloat64

	if c.Capital != nil {
		capital = sql.NullString{String: *c.Capital, Valid: true}
	}
	if c.Region != nil {
		region = sql.NullString{String: *c.Region, Valid: true}
	}
	if c.CurrencyCode != nil {
		currency = sql.NullString{String: *c.CurrencyCode, Valid: true}
	}
	if c.FlagURL != nil {
		flag = sql.NullString{String: *c.FlagURL, Valid: true}
	}
	if c.ExchangeRate != nil {
		exchange = sql.NullFloat64{Float64: *c.ExchangeRate, Valid: true}
	}
	if c.EstimatedGDP != nil {
		est = sql.NullFloat64{Float64: *c.EstimatedGDP, Valid: true}
	}

	_, err := tx.Exec(q,
		c.Name,
		capital,
		region,
		c.Population,
		currency,
		exchange,
		est,
		flag,
		c.LastRefreshedAt,
	)

	return err
}

// GetAll returns countries matching optional filters and sorting
func GetAll(db *sql.DB, region, currency, sort string) ([]Country, error) {
	base := `SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at FROM countries`
	var filters []interface{}
	where := ""
	if region != "" {
		where += " WHERE region = ?"
		filters = append(filters, region)
	}
	if currency != "" {
		if where == "" {
			where += " WHERE currency_code = ?"
		} else {
			where += " AND currency_code = ?"
		}
		filters = append(filters, currency)
	}

	order := ""
	if sort == "gdp_desc" {
		order = " ORDER BY estimated_gdp DESC"
	}

	q := base + where + order
	rows, err := db.Query(q, filters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Country
	for rows.Next() {
		var c Country
		var capital, region, currency, flag sql.NullString
		var exchange, est sql.NullFloat64
		var last sql.NullTime

		if err := rows.Scan(&c.ID, &c.Name, &capital, &region, &c.Population, &currency, &exchange, &est, &flag, &last); err != nil {
			return nil, err
		}
		if capital.Valid {
			c.Capital = &capital.String
		}
		if region.Valid {
			c.Region = &region.String
		}
		if currency.Valid {
			c.CurrencyCode = &currency.String
		}
		if exchange.Valid {
			c.ExchangeRate = &exchange.Float64
		}
		if est.Valid {
			c.EstimatedGDP = &est.Float64
		}
		if flag.Valid {
			c.FlagURL = &flag.String
		}
		if last.Valid {
			c.LastRefreshedAt = &last.Time
		}
		out = append(out, c)
	}

	return out, nil
}

// GetByName fetches a single country by case-insensitive name
func GetByName(db *sql.DB, name string) (*Country, error) {
	q := `SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at FROM countries WHERE LOWER(name) = LOWER(?) LIMIT 1`
	row := db.QueryRow(q, name)

	var c Country
	var capital, region, currency, flag sql.NullString
	var exchange, est sql.NullFloat64
	var last sql.NullTime

	if err := row.Scan(&c.ID, &c.Name, &capital, &region, &c.Population, &currency, &exchange, &est, &flag, &last); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if capital.Valid {
		c.Capital = &capital.String
	}
	if region.Valid {
		c.Region = &region.String
	}
	if currency.Valid {
		c.CurrencyCode = &currency.String
	}
	if exchange.Valid {
		c.ExchangeRate = &exchange.Float64
	}
	if est.Valid {
		c.EstimatedGDP = &est.Float64
	}
	if flag.Valid {
		c.FlagURL = &flag.String
	}
	if last.Valid {
		c.LastRefreshedAt = &last.Time
	}

	return &c, nil
}

// DeleteByName deletes a country by name
func DeleteByName(db *sql.DB, name string) (bool, error) {
	q := `DELETE FROM countries WHERE LOWER(name) = LOWER(?)`
	res, err := db.Exec(q, name)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// TotalCount returns number of countries
func TotalCount(db *sql.DB) (int64, error) {
	q := `SELECT COUNT(*) FROM countries`
	var n int64
	if err := db.QueryRow(q).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// SaveLastRefreshed stores the last refresh timestamp in metadata
func SaveLastRefreshed(tx *sql.Tx, t time.Time) error {
	q := `INSERT INTO metadata (meta_key, meta_value, updated_at) VALUES ('last_refreshed_at', ?, ?) ON DUPLICATE KEY UPDATE meta_value = VALUES(meta_value), updated_at = VALUES(updated_at)`
	_, err := tx.Exec(q, t.UTC().Format(time.RFC3339), t)
	return err
}

// GetLastRefreshed reads the last refresh timestamp
func GetLastRefreshed(db *sql.DB) (*time.Time, error) {
	q := `SELECT meta_value FROM metadata WHERE meta_key='last_refreshed_at' LIMIT 1`
	var v sql.NullString
	if err := db.QueryRow(q).Scan(&v); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !v.Valid || v.String == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v.String)
	if err != nil {
		// fallback to null
		return nil, nil
	}
	return &t, nil
}
