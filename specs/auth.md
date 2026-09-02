# Spec — Mòdul `auth`

## Context i abast d'aquesta versió (v1.1)

Suporta tres rols: **admin**, **teacher** (professor) i **student** (alumne). Aquest mòdul cobreix registre d'alumnes, login, logout, renovació de token (refresh) i protecció d'endpoints per rol.

**Decisions de disseny validades (2026-09-02)**:
- Autenticació basada en **par de tokens JWT**:
  - **Access Token (JWT)**: Visió stateless d'accés ràpid (15 minuts de validesa).
  - **Refresh Token**: Cadena aleatòria opaca guardada a la taula `refresh_tokens` de Postgres (7 dies de validesa).
  - **Taula de Revocació d'Access Tokens (`revoked_access_tokens`)**: En fer logout, el token d'accés actiu (identificat pel seu hash/jti) es registra a la taula de revocació a Postgres fins que expira. El middleware d'autenticació valida que l'Access Token no hagi estat revocat a la BD a cada petició protegida.
- El frontend guarda els tokens a `localStorage` i envia l'Access Token com a `Authorization: Bearer <token>` (via l'interceptor d'Axios centralitzat).
- Logout revoca tant el Refresh Token com l'Access Token actiu al servidor.

## Èpica 1: Login, registre i sessió

- Com a usuari (admin, teacher o student), vull fer login amb email i contrasenya per accedir a les funcionalitats del meu rol.
- Com a alumne, vull poder registrar-me des de la pàgina pública (`/auth/register`), obtenint automàticament el rol `student`.
- Com a usuari, vull rebre un missatge d'error clar si les credencials són incorrectes, sense revelar si l'error és l'email o la contrasenya.
- Com a sistema, vull emetre un par de tokens JWT (Access Token de 15 minuts + Refresh Token de 7 dies) en cada login/registre correcte.

## Èpica 2: Protecció d'endpoints per rol i verificació de revocació

- Com a sistema, vull un middleware (`AuthMiddleware`) que verifiqui la signatura HMAC del JWT i comprove que el token no figur com a revocat a la taula `revoked_access_tokens` a la BD. Retorna 401 si el token és invàlid, ha expirat o ha estat revocat.
- Com a sistema, vull poder marcar un endpoint amb `RequireRole("admin")`, `RequireRole("admin", "teacher")`, etc., retornant 403 Forbidden si el rol de l'usuari no és autoritzat.

## Èpica 3: Logout i revocació real

- Com a usuari, vull poder fer logout perquè tant el meu `refresh_token` com el meu `access_token` actiu quedin **realment invalidats al servidor** (`revoked_at` s'omple a `refresh_tokens` i el hash de l'Access Token es desaria a `revoked_access_tokens`), evitant que es puguin continuar fent peticions amb el token abans de la seva expiració.

## Èpica 4: Renovar tokens (Refresh)

- Com a usuari amb sessió activa o expirem-se l'Access Token, vull utilitzar el meu `refresh_token` per obtenir un nou par de tokens sense haver de reintroduir les meves credencials (`POST /auth/refresh`).

## Fora d'abast explícit
- Recuperació de contrasenya per email.
- Autenticació multifactor.
