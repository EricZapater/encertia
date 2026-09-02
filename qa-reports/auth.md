# QA — Mòdul auth

**Veredicte**: APTE
**Data**: 2026-09-02

## Compliment funcional
- [OK] Login i sessió JWT (Èpica 1) — El mòdul implementa l'esquema d'autenticació basat en par de tokens JWT (Access Token de 15m + Refresh Token de 7 dies guardat a `refresh_tokens`). El contracte `contracts/auth.openapi.yaml` ha estat actualitzat (v0.2.0) i homologat per a reflectir exactament les estructures `AuthResponse`, `TokenPair`, `/auth/register` i `/auth/refresh`.
- [OK] Missatge d'error de login ambigu (Èpica 1) — Rebutja credencials invàlides amb 401 Unauthorized sense revelar si la fallida prové de l'email o de la contrasenya.
- [OK] Middleware de verificació de sessió i revocació (Èpica 2) — `AuthMiddleware` valida la signatura HMAC del JWT i comprova que l'Access Token no figuri a la taula de revocació `revoked_access_tokens` a PostgreSQL, denegant l'accés immediatament (401) si ha estat invalidat.
- [OK] Protecció d'endpoints per rol (Èpica 2) — Implementat via middleware `RequireRole` per als rols `admin`, `teacher` i `student`.
- [OK] Logout efectiu al servidor (Èpica 3) — `/auth/logout` revoca el `refresh_token` a la base de dades i registra l'Access Token actiu a la taula de revocació `revoked_access_tokens`. Qualsevol petició posterior realitzada amb l'Access Token o Refresh Token anteriors és rebutjada amb `401 Unauthorized`.
- [OK] Registre públic d'alumnes i renovació de token (Èpica 4) — Endpoints `/auth/register` (assigna rol `student`) i `/auth/refresh` (rotació segura de refresh token) plenament operatius i coberts per tests.

## Qualitat de codi
- [OK] Compilació i tests — Execució neta de la suite completa de Go (`go test -count=1 ./...`) i tests frontend/type-check (`pnpm run type-check && pnpm run test`) 100% verda.
- [OK] Seguretat i hashing — Hashing de contrasenyes amb `bcrypt` i hashing SHA-256 (`hashToken`) per a tokens de revocació i refresh tokens a la base de dades.
- [OK] Validació d'entrada i gestió d'errors — Validació d'email, longitud de contrasenya (mínim 8 caràcters) i estructures d'error estandarditzades amb `ErrorResponse`.

## Homogeneïtat
- [OK] Estructura de carpetes backend — Segueix `internal/auth` amb `handler.go`, `service.go`, `repository.go`, `model.go` en singular, segons la `constitution.md`.
- [OK] Estructura de carpetes frontend — Segueix `src/modules/auth` en plural (`views/`, `store.ts`, `api.ts`, `types.ts`), segons la `constitution.md`.
- [OK] Nomenclatura dels rols d'usuari — Homologada l'enum de rols a `[admin, teacher, student]` entre el contracte `contracts/auth.openapi.yaml`, l'especificació `specs/auth.md` i la implementació Go/TS.

## Incidències
Cap incidència detectada. La taula de revocació `revoked_access_tokens` (migració `000007_create_revoked_access_tokens.up.sql`) resol plenament el problema d'invalidació real al logout amb JWT.
