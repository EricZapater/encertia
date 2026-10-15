-- Migració per afegir el payload (petició) i response_body (resposta) a audit_logs
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS payload TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS response_body TEXT;
