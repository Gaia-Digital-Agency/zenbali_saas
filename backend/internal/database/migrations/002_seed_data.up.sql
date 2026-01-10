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
