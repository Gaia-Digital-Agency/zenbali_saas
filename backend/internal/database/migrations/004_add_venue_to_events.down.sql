-- ===========================================
-- Remove venue field from events table
-- ===========================================

ALTER TABLE events
DROP COLUMN IF EXISTS venue;
