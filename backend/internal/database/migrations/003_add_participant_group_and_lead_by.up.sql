-- ===========================================
-- Add participant_group_type and lead_by fields to events table
-- ===========================================

ALTER TABLE events
ADD COLUMN participant_group_type VARCHAR(50),
ADD COLUMN lead_by VARCHAR(255);
