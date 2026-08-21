# Changelog

Format basat en [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versionat amb [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.0] - 2026-08-21
### Added
- **Contracte d'API**: Especificació funcional ([`contracts/user.spec.md`](file:///Users/eric.zapater/Developer/encertia/contracts/user.spec.md)) i contracte OpenAPI 3.0 ([`contracts/user.openapi.yaml`](file:///Users/eric.zapater/Developer/encertia/contracts/user.openapi.yaml)) per al mòdul d'usuaris (`user`).
- **Rols i RBAC**:
  - Incorporació del rol `admin` (superusuari) amb accés total.
  - Rol `teacher` amb capacitats de gestió d'alumnes (`student`).
  - Restricció de l'autoregistre públic (`/auth/register`) exclusivament per a `student`.
- **Backend (`internal/user`)**:
  - Migracions SQL (`000002_add_admin_role_and_user_features.up.sql` i `.down.sql`) amb columna `is_active`, check constraint actualitzat i nous índexs.
  - Endpoints complets: `GET /users` (paginació i filtres), `POST /users` (alta individual), `POST /users/batch` (alta massiva d'alumnes), `GET /users/:id`, `PUT /users/:id` (autoedició de perfil per usuaris actius i edició administrativa per a admins), `POST /users/:id/password` (reseteig administratiu de clau), `DELETE /users/:id` (soft-delete amb revocació de tokens).
  - Middleware de control de rols `RequireRole` a `internal/shared/`.
  - Suite exhaustiva de tests unitaris i d'integració a `service_test.go` i `handler_test.go` amb 100% de tests verds.
- **Frontend (`src/modules/users`)**:
  - Tipus TypeScript del contracte, client d'API Axios (`api.ts`) i store Pinia (`useUserStore`).
  - Utilitat de parseig client-side de fitxers CSV/TSV (`csvParser.ts`) amb detecció automàtica de delimitador, validació per fila i suport de capçaleres en català/anglès.
  - Vistes completes: taula paginada amb filtres (`UsersListView.vue`), formulari modal d'alta/edició amb control de rols (`UserFormModal.vue`), diàleg de reseteig de contrasenya (`ResetPasswordModal.vue`), importador massiu CSV en 3 passos amb previsualització i informe d'errors (`BatchImportModal.vue`), i vista de detall (`UserDetailView.vue`).
  - Protecció de rutes amb navigation guards segons rols (`admin`, `teacher`).
  - Suite de proves unitàries i de components a `__tests__/`.
- **Infraestructura i CI/CD**:
  - Dockerfile multi-stage per a Backend (Go compilat estàtic sobre Alpine amb usuari no root) i Frontend (Vite/pnpm compilat servit per Nginx amb suport de routing SPA).
  - Configuració de desenvolupament local amb `docker-compose.yml` (PostgreSQL 16 amb healthcheck i volum persistent, Backend i Frontend) i plantilla `.env.example`.
  - Pipelines de GitHub Actions: `.github/workflows/ci.yml` (tests i builds automatitzats de Go i Vue) i `.github/workflows/deploy.yml` (generació d'imatges i preparació de CD).
### Added
- **Contracte d'API**: Especificació funcional (`contracts/auth.spec.md`) i contracte OpenAPI 3.0 (`contracts/auth.openapi.yaml`) per al mòdul d'autenticació.
- **Backend (`internal/auth`)**:
  - Implementació modular en Go/Gin amb PostgreSQL i SQL pur sense ORM (handler, service, repository, model).
  - Endpoints d'autenticació: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`.
  - Hashing segur de contrasenyes amb bcrypt i emissió de tokens JWT (HMAC-SHA256) amb refresh tokens persistits.
  - Migracions SQL (`internal/db/migrations/000001_create_auth_tables.up.sql` i `.down.sql`) amb suport per soft-delete.
  - Middleware transversal d'autenticació JWT i CORS a `internal/shared/`.
  - Suite de proves unitàries i d'integració amb 100% de cobertura satisfactòria.
- **Frontend (`src/modules/auth`)**:
  - Configuració base de Vue 3 (Composition API), Vite, TypeScript, PrimeVue, Pinia i Vue Router.
  - Client Axios centralitzat (`src/api/client.ts`) amb interceptors per injecció de Bearer token i renovació transparent (auto-refresh) en rebre 401.
  - Store Pinia (`useAuthStore`) amb gestió d'estat d'autenticació i persistència a `localStorage`.
  - Vistes d'inici de sessió (`LoginView.vue`), registre (`RegisterView.vue`) amb selecció de rol (professor/alumne), i panell de perfil (`ProfileView.vue`).
  - Navigation guards a Vue Router per protegir rutes autenticades.

## [0.1.0] - 2026-08-21
### Added
- Estructura inicial del projecte: constitution, rols d'agent
  (orquestrador, backend, frontend, infra), esquelet de carpetes.
