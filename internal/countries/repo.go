package countries

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/zjoart/countryxchange/pkg/logger"
)

var ErrNotFound = errors.New("not found")

// DropTables drops the countries and metadata tables
func DropTables(db *sql.DB) error {
	logger.Info("repo: DropTables start")

	// Drop tables in reverse order of dependencies
	dropMetadata := `DROP TABLE IF EXISTS metadata;`
	if _, err := db.Exec(dropMetadata); err != nil {
		logger.Error("repo: drop metadata table failed", logger.WithError(err))
		return err
	}

	dropCountries := `DROP TABLE IF EXISTS countries;`
	if _, err := db.Exec(dropCountries); err != nil {
		logger.Error("repo: drop countries table failed", logger.WithError(err))
		return err
	}

	logger.Info("repo: DropTables complete")
	return nil
}

// EnsureTables creates countries and metadata tables when needed
func EnsureTables(db *sql.DB) error {
	logger.Info("repo: EnsureTables start")
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
		logger.Error("repo: create countries table failed", logger.WithError(err))
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
		logger.Error("repo: create metadata table failed", logger.WithError(err))
		return err
	}

	logger.Info("repo: EnsureTables complete")
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

	if err != nil {
		logger.Error("repo: UpsertCountry failed", logger.Fields{"country": c.Name}, logger.WithError(err))
	}
	return err
}

// GetAll returns countries matching optional filters and sorting
func GetAll(db *sql.DB, region, currency, sort string) ([]Country, error) {
	base := `SELECT id, name, capital, region, population, currency_code, exchange_rate, estimated_gdp, flag_url, last_refreshed_at FROM countries`

	// Build WHERE conditions in a slice so multiple filters combine cleanly
	var conds []string
	var args []interface{}
	if region != "" {
		// case-insensitive match
		conds = append(conds, "LOWER(region) = LOWER(?)")
		args = append(args, region)
	}
	if currency != "" {
		conds = append(conds, "LOWER(currency_code) = LOWER(?)")
		args = append(args, currency)
	}

	order := ""
	if sort == "gdp_desc" {
		order = " ORDER BY estimated_gdp DESC"
	} else if sort == "gdp_asc" {
		order = " ORDER BY estimated_gdp ASC"
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	q := base + where + order
	logger.Debug("repo: GetAll final query", logger.Fields{"query": q, "args": args})
	rows, err := db.Query(q, args...)
	if err != nil {
		logger.Error("repo: GetAll query failed", logger.WithError(err))
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

	logger.Info("repo: GetAll complete", logger.Fields{"count": len(out)})
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
			logger.Debug("repo: GetByName not found", logger.Fields{"name": name})
			return nil, ErrNotFound
		}
		logger.Error("repo: GetByName failed", logger.Fields{"name": name}, logger.WithError(err))
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

	logger.Info("repo: GetByName success", logger.Fields{"name": c.Name, "id": c.ID})
	return &c, nil
}

// DeleteByName deletes a country by name
func DeleteByName(db *sql.DB, name string) (bool, error) {
	q := `DELETE FROM countries WHERE LOWER(name) = LOWER(?)`
	res, err := db.Exec(q, name)
	if err != nil {
		logger.Error("repo: DeleteByName failed", logger.Fields{"name": name}, logger.WithError(err))
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		logger.Error("repo: DeleteByName RowsAffected failed", logger.Fields{"name": name}, logger.WithError(err))
		return false, err
	}
	logger.Info("repo: DeleteByName result", logger.Fields{"name": name, "deleted": n > 0})
	return n > 0, nil
}

// TotalCount returns number of countries
func TotalCount(db *sql.DB) (int64, error) {
	q := `SELECT COUNT(*) FROM countries`
	var n int64
	if err := db.QueryRow(q).Scan(&n); err != nil {
		logger.Error("repo: TotalCount failed", logger.WithError(err))
		return 0, err
	}
	logger.Info("repo: TotalCount", logger.Fields{"count": n})
	return n, nil
}

// SaveLastRefreshed stores the last refresh timestamp in metadata
func SaveLastRefreshed(tx *sql.Tx, t time.Time) error {
	q := `INSERT INTO metadata (meta_key, meta_value, updated_at) VALUES ('last_refreshed_at', ?, ?) ON DUPLICATE KEY UPDATE meta_value = VALUES(meta_value), updated_at = VALUES(updated_at)`
	_, err := tx.Exec(q, t.UTC().Format(time.RFC3339), t)
	if err != nil {
		logger.Error("repo: SaveLastRefreshed failed", logger.WithError(err))
	} else {
		logger.Info("repo: SaveLastRefreshed", logger.Fields{"t": t.UTC().Format(time.RFC3339)})
	}
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
		logger.Error("repo: GetLastRefreshed failed", logger.WithError(err))
		return nil, err
	}
	if !v.Valid || v.String == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v.String)
	if err != nil {
		// fallback to null
		logger.Warn("repo: GetLastRefreshed parse failed", logger.WithError(err))
		return nil, nil
	}
	logger.Info("repo: GetLastRefreshed", logger.Fields{"t": t.UTC().Format(time.RFC3339)})
	return &t, nil
}
