# Especificació funcional — Mòdul Mètriques, Monitoratge i Auditoria (`metrics`) (v1.2.0)

**Versió**: 1.2.0  
**Estat**: En revisió (Pendent de validació humana - Checkpoint 1)  
**Data**: 2026-09-02  
**Accés**: Restringit **exclusivament al rol `admin`**  

---

## 1. Objectiu del Mòdul

El mòdul **Mètriques (`metrics`)** proporciona a l'Administrador del sistema una visió integral en temps real de l'activitat de la plataforma Encertia, el rendiment de l'API REST/WebSockets, el registre d'auditoria d'ús per usuari i la salut del servidor i de la base de dades PostgreSQL.

---

## 2. Èpiques i Històries d'Usuari

### Èpica 1: Registre d'Ús i Auditoria (Qui fa servir què)
- **HU-MET-01 (Audit Log d'accions d'usuari)**: Enregistrament automàtic de cada petició/acció rellevant realitzada a l'aplicació, guardant `user_id`, `user_email`, `user_role`, `action`, `module`, `endpoint`, `ip_address`, `status_code` i `timestamp`.
- **HU-MET-02 (Cerca i filtre d'auditoria)**: L'Administrador pot cercar i filtrar el registre d'activitat per usuari, per mòdul (`auth`, `users`, `quizzes`, `courses`, `materials`, `match`, `evaluations`), per data o per codi d'estat HTTP.
- **HU-MET-03 (Exportació CSV d'Auditoria)**: Botó per descarregar el registre d'auditoria en format CSV per a anàlisi externa o compliment normatiu.

### Èpica 2: Mètriques de Rendiment de l'API (Rendiment i Latència)
- **HU-MET-04 (Temps de resposta dels endpoints)**: Captura mitjançant Middleware de Gin del temps de resposta en mil·lilisegons (`duration_ms`) de cada endpoint de l'API.
- **HU-MET-05 (Indicadors clau de latència)**: Càlcul automàtic de la durada mitjana (`avg_duration_ms`), p95 (`95th percentile`) i p99 (`99th percentile`), així com el total de peticions per segon (RPS) i la taxa d'errors (HTTP 4xx i 5xx).
- **HU-MET-06 (Rànquing d'endpoints més lents i amb més errors)**: Llistat interactiu de les rutes que presenten major latència o major taxa d'error.

### Èpica 3: Proposta d'Ampliació de Mètriques (Propostes Addicionals)
- **HU-MET-07 (Mètriques de Partides i Gamificació)**: Total de partides jugades (`matches_played`), alumnes actius en temps real, mitjana d'alumnes per partida i mitjana de respostes encertades.
- **HU-MET-08 (Mètriques d'Engagement de contingut)**: Total de visualitzacions/lectures de PDF, reproduccions de vídeo i cursos més actius.
- **HU-MET-09 (Salut del Sistema i BD)**: Uptime del servidor Go, nombre de goroutines actives, ús de memòria RAM i estat del pool de connexions a PostgreSQL (`open_connections`, `in_use`).

---

## 3. Normes de Negoci i Restriccions

1. **Restricció d'Accés RBAC**: Únicament usuaris amb rol `admin` tenen accés als endpoints `/api/v1/metrics/*` i a la vista `/metrics` del frontend. Intentar accedir amb rol `teacher` o `student` retorna `403 Forbidden`.
2. **Middleware Asíncron i no Bloquejant**: L'enregistrament de mètriques i audit logs s'executa en segon pla (goroutines a Go) per no afegir cap latència a les peticions dels usuaris finals.
3. **Conservació de Dades**: Els registres detallats d'auditoria es conserven a la taula `audit_logs` de PostgreSQL.

---

## 4. Esquema de Dades Relacional (PostgreSQL)

```sql
-- Migració 000011: Creació de taules de mètriques i registre d'auditoria
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255),
    user_role VARCHAR(50),
    action VARCHAR(100) NOT NULL,
    module VARCHAR(50) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INT NOT NULL,
    duration_ms INT NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_module ON audit_logs(module);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```
