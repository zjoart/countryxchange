-- Create database and schema for Country Xchange
-- Create tables for Country Xchange (creates tables in the currently connected database)

-- Create countries table
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 4) Create metadata table (used to store last_refreshed_at)
CREATE TABLE IF NOT EXISTS metadata (
  meta_key VARCHAR(128) PRIMARY KEY,
  meta_value VARCHAR(1024),
  updated_at DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
