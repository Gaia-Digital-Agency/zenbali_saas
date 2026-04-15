-- ===========================================
-- Rollback participant_group_type and lead_by fields from events table
-- ===========================================

ALTER TABLE events
DROP COLUMN IF EXISTS participant_group_type,
DROP COLUMN IF EXISTS lead_by;
