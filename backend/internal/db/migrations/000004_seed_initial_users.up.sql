-- Inserció idempotent dels usuaris inicials
INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
VALUES 
  (gen_random_uuid(), 'hola@ericzapater.cat', '$2a$10$gLjxxFqNpASy0.V3h1bXpOel8H3G4Wa/P.yvJp5LfJUGhFQDAqxfi', 'Èric', 'Zapater', 'admin', true, NOW(), NOW()),
  (gen_random_uuid(), 'dgarage21@gmail.com', '$2a$10$bgMAkGEWZdqTRG64iLb7JuY2zn89g6YAqjs7KIJbnvudLesR53JZ.', 'Albert', 'Caballer', 'teacher', true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
