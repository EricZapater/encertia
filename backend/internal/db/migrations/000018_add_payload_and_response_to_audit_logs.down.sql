-- Revertir l'addició de payload i response_body a audit_logs
ALTER TABLE audit_logs DROP COLUMN IF EXISTS response_body;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS payload;
