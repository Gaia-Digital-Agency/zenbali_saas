-- ===========================================
-- Add venue field to events table
-- ===========================================

ALTER TABLE events
ADD COLUMN venue VARCHAR(50);
