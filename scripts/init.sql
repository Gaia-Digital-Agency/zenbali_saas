-- Zen Bali Database Initialization Script
-- This script runs when the PostgreSQL container starts

-- Create the database (if not exists is handled by Docker)
-- The database is created by POSTGRES_DB environment variable

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE zenbali TO zenbali;

-- The main schema will be created by Go migrations
