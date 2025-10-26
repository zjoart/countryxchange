# Country Xchange

Country Currency & Exchange API

This service fetches country data from the REST Countries API, fetches exchange rates, computes an estimated GDP and caches the data in a MySQL database. It exposes REST endpoints to refresh the cache and query country data.

## Quick overview

Endpoints

- POST /countries/refresh — Fetch countries and exchange rates, then cache them
- GET /countries — List countries (filters: `?region=...`, `?currency=...`, `?sort=gdp_desc`)
- GET /countries/:name — Get a country by name (case-insensitive)
- DELETE /countries/:name — Delete a country
- GET /status — Show total countries and last refresh timestamp
- GET /countries/image — Serve generated summary image (cache/summary.png)

All responses are JSON unless noted (image endpoint).

## Config / .env

The project uses environment variables. Create a `.env` at project root with at least:

```
PORT=8080
APP_ENV=development
DB_USER=root
DB_PASS=secret
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=countryxchange
API_BASE=localhost:8080
SWAGGER_SCHEMES=http
```

## Database

The service uses MySQL. The code will create required tables automatically when refreshing. Ensure the database specified by `DB_NAME` exists and the user has privileges.

Schema created by the app (automatically):
- `countries` table — stores country records
- `metadata` table — stores last refresh timestamp

## How it works

1. `POST /countries/refresh` fetches all countries and the USD exchange rates. For each country:
   - uses the first currency from the country's currencies array
   - looks up its exchange rate from the exchange API
   - computes `estimated_gdp = population * random(1000-2000) / exchange_rate`
   - stores or updates the DB record (matching by name, case-insensitive)
   - if currencies array is empty, currency_code/exchange_rate set to null and estimated_gdp set to 0
   - if currency not found in rates, exchange_rate and estimated_gdp are null
2. After a successful refresh the service saves a `last_refreshed_at` timestamp and generates `cache/summary.png` containing total countries, top 5 by estimated GDP and timestamp.

If either external API fails the refresh will abort and return 503 — no DB changes are made.

## Run locally

Make sure `go` is installed and `go` modules can download dependencies.

```bash
# from project root
go mod tidy
go build ./cmd/app
./app
```

Alternatively use `go run`:

```bash
go run ./cmd/app
```

## Notes & next steps

- The image generator uses a simple library to draw the summary PNG at `cache/summary.png`.
- You can improve fonts, layout and caching strategy as needed.
