-- ===========================================
-- Zen Bali Migration Rollback
-- ===========================================

-- Drop triggers
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP TRIGGER IF EXISTS update_admins_updated_at ON admins;
DROP TRIGGER IF EXISTS update_creators_updated_at ON creators;
DROP TRIGGER IF EXISTS update_entrance_types_updated_at ON entrance_types;
DROP TRIGGER IF EXISTS update_event_types_updated_at ON event_types;
DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS visitors;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS creators;
DROP TABLE IF EXISTS entrance_types;
DROP TABLE IF EXISTS event_types;
DROP TABLE IF EXISTS locations;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
