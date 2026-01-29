-- Initial Admin Account (password: admin123)
INSERT INTO admins (email, password_hash, name, is_active) VALUES
('admin@zenbali.org', '$2a$10$7E1R11SN79ghwafaxYckH./LZWgy4TagcjIgMJE94ByW5WACFTCD.', 'Admin', true)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;