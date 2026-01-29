-- ===========================================
-- Zen Bali Seed Data Migration
-- ===========================================

-- ===========================================
-- Locations (25 Bali Areas)
-- ===========================================

INSERT INTO locations (name, slug) VALUES
    ('Ubud', 'ubud'),
    ('Canggu', 'canggu'),
    ('Seminyak', 'seminyak'),
    ('Kuta', 'kuta'),
    ('Legian', 'legian'),
    ('Sanur', 'sanur'),
    ('Nusa Dua', 'nusa-dua'),
    ('Uluwatu', 'uluwatu'),
    ('Jimbaran', 'jimbaran'),
    ('Denpasar', 'denpasar'),
    ('Tabanan', 'tabanan'),
    ('Gianyar', 'gianyar'),
    ('Karangasem', 'karangasem'),
    ('Singaraja', 'singaraja'),
    ('Lovina', 'lovina'),
    ('Amed', 'amed'),
    ('Candidasa', 'candidasa'),
    ('Padang Bai', 'padang-bai'),
    ('Munduk', 'munduk'),
    ('Bedugul', 'bedugul'),
    ('Tegallalang', 'tegallalang'),
    ('Sidemen', 'sidemen'),
    ('Nusa Penida', 'nusa-penida'),
    ('Nusa Lembongan', 'nusa-lembongan'),
    ('Kintamani', 'kintamani')
ON CONFLICT (slug) DO NOTHING;

-- ===========================================
-- Event Types (25 Types)
-- ===========================================

INSERT INTO event_types (name, slug) VALUES
    ('Yoga', 'yoga'),
    ('Healing', 'healing'),
    ('Therapy', 'therapy'),
    ('Show', 'show'),
    ('Theater', 'theater'),
    ('Music Concert', 'music-concert'),
    ('Dance Performance', 'dance-performance'),
    ('Art Exhibition', 'art-exhibition'),
    ('Workshop', 'workshop'),
    ('Retreat', 'retreat'),
    ('Meditation', 'meditation'),
    ('Sound Healing', 'sound-healing'),
    ('Breathwork', 'breathwork'),
    ('Ecstatic Dance', 'ecstatic-dance'),
    ('Festival', 'festival'),
    ('Market & Bazaar', 'market-bazaar'),
    ('Food & Culinary', 'food-culinary'),
    ('Sports & Fitness', 'sports-fitness'),
    ('Wellness', 'wellness'),
    ('Spiritual Ceremony', 'spiritual-ceremony'),
    ('Photography', 'photography'),
    ('Film Screening', 'film-screening'),
    ('Comedy', 'comedy'),
    ('Networking', 'networking'),
    ('Community Gathering', 'community-gathering')
ON CONFLICT (slug) DO NOTHING;

-- ===========================================
-- Entrance Types (6 Types)
-- ===========================================

INSERT INTO entrance_types (name, slug) VALUES
    ('Free', 'free'),
    ('Prepaid Online', 'prepaid-online'),
    ('Pay at Site', 'pay-at-site'),
    ('Donation-based', 'donation-based'),
    ('By Registration Only', 'registration-only'),
    ('Members Only', 'members-only')
ON CONFLICT (slug) DO NOTHING;

-- ===========================================
-- Admin User (password: admin123)
-- ===========================================

INSERT INTO admins (email, password_hash, name, is_active) VALUES
    ('admin@zenbali.org', '$2a$10$7E1R11SN79ghwafaxYckH./LZWgy4TagcjIgMJE94ByW5WACFTCD.', 'Admin', true)
ON CONFLICT (email) DO NOTHING;

-- ===========================================
-- Test Creator (password: admin123)
-- ===========================================

INSERT INTO creators (id, name, organization_name, email, mobile, password_hash, is_verified, is_active) VALUES
    ('98d34804-bf30-49b1-b578-af23b6b7c124', 'Test Creator', 'Test Organization', 'creator@test.com', '+628123456789', '$2a$10$7E1R11SN79ghwafaxYckH./LZWgy4TagcjIgMJE94ByW5WACFTCD.', true, true)
ON CONFLICT (email) DO NOTHING;

-- ===========================================
-- Sample Paid & Published Event
-- ===========================================

INSERT INTO events (creator_id, title, event_date, event_time, location_id, event_type_id, duration, entrance_type_id, entrance_fee, contact_email, contact_mobile, notes, is_paid, is_published)
SELECT
    '98d34804-bf30-49b1-b578-af23b6b7c124'::uuid,
    'Sample Yoga Session in Ubud',
    CURRENT_DATE + INTERVAL '1 day',
    '09:00:00',
    (SELECT id FROM locations WHERE slug = 'ubud'),
    (SELECT id FROM event_types WHERE slug = 'yoga'),
    '2 hours',
    (SELECT id FROM entrance_types WHERE slug = 'prepaid-online'),
    150000,
    'creator@test.com',
    '+628123456789',
    'This is a sample yoga session for testing purposes.',
    true,
    true
WHERE NOT EXISTS (SELECT 1 FROM events WHERE title = 'Sample Yoga Session in Ubud');
