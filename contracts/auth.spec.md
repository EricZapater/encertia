# Especificació del Mòdul: Auth

## 1. Visió General
El mòdul d'autenticació (`auth`) gestiona la identitat, registre, inici de sessió, renovació de credencials i context de l'usuari actual a la plataforma Encertia.

Encertia és una plataforma educativa (LMS + Quiz) amb dos rols principals d'usuari:
- **`teacher`** (Professor): Gestiona cursos, continguts, sessions i qüestionaris.
- **`student`** (Alumne): Accedeix a cursos matriculats, consumeix continguts i respon qüestionaris.

## 2. Històries d'Usuari

### HU-AUTH-01: Registre d'Usuaris
- **Com a** nou usuari (alumne o professor).
- **Vull** registrar-me introduint nom, cognoms, correu electrònic i contrasenya.
- **Per tal de** tenir un compte a Encertia i accedir als serveis de la plataforma.
- **Criteris d'acceptació**:
  - El correu ha de ser únic i vàlid.
  - La contrasenya ha de tenir com a mínim 8 caràcters.
  - Si el correu ja existeix, retorna un error `409 Conflict` amb el codi `EMAIL_ALREADY_EXISTS`.
  - En registrar-se amb èxit (`201 Created`), es retorna el perfil creat i els tokens d'accés inicials.

### HU-AUTH-02: Inici de Sessió
- **Com a** usuari registrat.
- **Vull** iniciar sessió amb el meu correu electrònic i contrasenya.
- **Per tal d'** obtenir les credencials d'accés (JWT) per consumir l'API.
- **Criteris d'acceptació**:
  - Credencials correctes retornen `200 OK` amb `accessToken`, `refreshToken` i les dades bàsiques de l'usuari.
  - Credencials incorrectes retornen `401 Unauthorized` amb el codi `INVALID_CREDENTIALS` (sense revelar si l'error és el correu o la contrasenya).

### HU-AUTH-03: Renovació de Token (Refresh)
- **Com a** aplicació client frontend (Axios interceptor).
- **Vull** sol·licitar un nou `accessToken` utilitzant un `refreshToken` vigent.
- **Per tal de** mantenir la sessió activa de forma transparent sense demanar re-login a l'usuari constantment.
- **Criteris d'acceptació**:
  - Un `refreshToken` vàlid i no revocat retorna un nou parell de tokens.
  - Un `refreshToken` expirat o revocat retorna `401 Unauthorized` amb `TOKEN_EXPIRED` o `INVALID_TOKEN`.

### HU-AUTH-04: Tancament de Sessió (Logout)
- **Com a** usuari autenticat.
- **Vull** tancar la sessió.
- **Per tal d'** invalidar el meu `refreshToken` i assegurar que ningú pugui usar la meva sessió.
- **Criteris d'acceptació**:
  - Retorna `200 OK` i invalida el `refreshToken` a la base de dades.

### HU-AUTH-05: Obtenir Perfil Actual (Me)
- **Com a** usuari autenticat.
- **Vull** consultar les meves dades d'usuari i rol actual (`GET /auth/me`).
- **Per tal de** rehidratar l'estat del frontend (Pinia store) en carregar l'aplicació.
- **Criteris d'acceptació**:
  - Requereix capçalera `Authorization: Bearer <accessToken>`.
  - Retorna dades de l'usuari (`id`, `email`, `firstName`, `lastName`, `role`, `createdAt`).

## 3. Decisions Tècniques i Trade-offs
1. **Estratègia d'Autenticació: JWT Bearer Token + Refresh Token**:
   - `accessToken`: JWT de curta durada (ex. 15 minuts) firmat amb clau secreta (HMAC-SHA256), que conté `userId` i `role`.
   - `refreshToken`: Token opac o JWT de llarga durada (ex. 7 dies) persistit a la base de dades per permetre revocació.
   - *Trade-off*: S'utilitza Bearer Header per Axios per facilitar el consum client/API i separar clarament el client SPA del backend Gin, gestionant el cicle de vida del token mitjançant interceptors a `frontend/src/api/client.ts`.
2. **Estructura d'Errors Unificada**:
   - Tots els errors retornen JSON amb schema: `{ "error": { "code": "STRING_CODE", "message": "Missatge llegible", "details": {} } }`.
3. **Nomenclatura**:
   - JSON: `camelCase` (ex. `firstName`, `accessToken`, `expiresIn`).
   - PostgreSQL (a la capa DB): `snake_case` (ex. `first_name`, `refresh_tokens`).
