# Country Xchange

Country Currency & Exchange API

The service fetches country data from the REST Countries API, fetches exchange rates, computes an estimated GDP and caches the data in a MySQL database. It exposes REST endpoints to refresh the cache and query country data.

## Live API
The API is deployed and available at:
```
https://exciting-gratitude-production.up.railway.app
```

View the complete API documentation at:
```
https://exciting-gratitude-production.up.railway.app/swagger/index.html
```

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

The project uses environment variables. Copy the `.env.example` file to `.env` at the project root and update the values as needed:

```bash
cp .env.example .env
```

Then edit the `.env` file with your configuration values.

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

You can run the service locally for development. The production version is deployed at `https://exciting-gratitude-production.up.railway.app`.

Make sure `go` is installed and `go` modules can download dependencies. This project includes a `Makefile` with convenient targets. From the project root you can use:

```bash
# tidy modules (same as 'go mod tidy')
make tidy

# run the application (same as 'go run ./cmd/app')
make run

# show available make targets
make help
```

If you prefer the raw go commands, the equivalent are:

```bash
go mod tidy
go run ./cmd/app
```

### Database Setup

The service will automatically create the required database tables (`countries` and `metadata`) when you call `POST /countries/refresh` for the first time. The tables are created using `CREATE TABLE IF NOT EXISTS` statements. Just ensure that:

1. The MySQL database specified in your `.env` (`DB_NAME`) exists
2. The configured database user has sufficient privileges to create tables

## API Documentation

The API is documented using Swagger/OpenAPI. You can access the interactive API documentation when the service is running:

```bash
# Access Swagger UI locally
http://localhost:8080/swagger/index.html

# Access deployed Swagger UI
https://exciting-gratitude-production.up.railway.app/swagger/index.html
```

The Swagger UI provides:
- Interactive API documentation
- Request/response schemas
- Example values
- Try-it-out functionality to test endpoints directly
- Detailed error responses
- Query parameter documentation

## Notes & next steps

- The image generator uses a simple library to draw the summary PNG at `cache/summary.png`.
- You can improve fonts, layout and caching strategy as needed.
