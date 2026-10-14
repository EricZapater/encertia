# QA — Mòdul Autorització de Dispositius (`device_auth` + `agent`)

**Veredicte**: APTE
**Data**: 2026-10-14

## Entorn de proves
- **Backend compilat i tests unitaris executats**: SÍ
  - `go test -v ./internal/device/...` → **PASS** (14/14 tests passats, 0.51s)
  - `go test -v ./internal/auth/...` → **PASS** (16/16 tests passats incloent autorització de dispositiu per a docents, 1.74s)
- **Agent Go compilat i tests executats**: SÍ
  - `go test -v ./...` a `agent/` → **PASS** (`TestGetHardwareUUID`, `TestGetHardwareUUIDHash`)
  - Compilació multiplataforma (`bash build.sh`) generant binaris autònoms a `agent/bin/`
  - Verificació d'extracció de maquinari real en local mitjançant `./bin/encertia-agent-darwin --debug`
- **Frontend compilat i tests executats**: SÍ
  - `pnpm test` a `frontend/` → **PASS** (31/31 fitxers de test, 169/169 tests passats, 3.76s)
  - `pnpm build` a `frontend/` → **OK** (`vue-tsc -b && vite build` completat en 1.91s sense cap error de tipus TypeScript)
- **Base de dades i migració**: Esquema actualitzat amb la migració `000016_create_authorized_devices.up.sql` (`authorized_devices` i `device_bootstrap_tokens`) i la corresponent `000016_create_authorized_devices.down.sql`.
- **Dades de prova generades**: Coberta íntegrament mitjançant repositoris mock en memòria en els tests unitaris i d'integració de Go i Vitest.

## Proves funcionals executades
- **Verificació de dispositiu natiu (`POST /devices/verify`)**:
  - Hash SHA-256 hexadecimal vàlid d'equip registrat i actiu → Retorna 200 OK amb `bootstrapToken` (prefix `dbt_`), `teacherEmail`, `deviceName` i `expiresInSeconds: 60`.
  - Hash d'equip no registrat o desactivat → Retorna 403 Forbidden amb codi d'error `DEVICE_NOT_AUTHORIZED`.
  - Hash malformat o no hexadecimal → Retorna 400 Bad Request amb codi `VALIDATION_ERROR`.
- **Gestió administrativa d'equips (`/admin/devices`)**:
  - `POST /admin/devices` (amb rol `admin`) → Alta i reactivació d'equips associats a un docent existent (`201 Created`).
  - `POST /admin/devices` per usuaris no administradors (`teacher`/`student`) → `403 Forbidden`.
  - `POST /admin/devices` amb correu de docent inexistent o compte desactivat → `400 Bad Request`.
  - `GET /admin/devices` (amb rol `admin`) → Llistat de tots els dispositius registrats amb dades del docent i última connexió (`200 OK`).
  - `DELETE /admin/devices/:id` (amb rol `admin`) → Revocació/desactivació immediata de l'ordinador (`204 No Content`).
- **Autenticació docent amb token de dispositiu (`POST /auth/login`)**:
  - Docent amb `deviceToken` vàlid emès per l'equip assignat → Login completat amb èxit (200 OK), el token es marca com a consumit i s'actualitza el timestamp `last_used_at` de l'equip.
  - Reutilització d'un `deviceToken` ja consumit o expirat → Bloquejat amb 403 Forbidden (`DEVICE_AUTHORIZATION_REQUIRED`).
  - `deviceToken` utilitzat per un docent diferent al vinculat a l'equip → Bloquejat amb 403 Forbidden (`DEVICE_AUTHORIZATION_REQUIRED`).
  - Mode suau (Fase 1, `ENFORCE_TEACHER_DEVICE_AUTH=false`): docent sense token → Login permès amb normalitat per evitar disrupcions d'aula.
  - Mode estricte (Fase 2, `ENFORCE_TEACHER_DEVICE_AUTH=true`): docent sense token → Bloquejat amb 403 Forbidden (`DEVICE_AUTHORIZATION_REQUIRED`).
  - Alumnes (`student`) i administradors (`admin`) → Login permès amb credencials habituals independentment de si s'envia o no `deviceToken`.
- **Execució i extracció d'identificador de maquinari a l'Agent Go (`agent/`)**:
  - Extracció del UUID de la placa base (SMBIOS UUID) provat en macOS via `ioreg -rd1 -c IOPlatformExpertDevice`.
  - Implementat amb suport per a Windows (PowerShell CIM instance `Win32_ComputerSystemProduct` i fallback `wmic`) i Linux (`/sys/class/dmi/id/product_uuid` i fallback `/etc/machine-id`).
  - Càlcul del hash SHA-256 hexadecimal en minúscules sense espais.
  - Generació d'URL segura cap a `/login?deviceToken=...&email=...` i obertura del navegador per defecte del sistema operatiu (`browser.OpenURL`).
  - Assistent interactiu per consola (`--register` / `-r`) guiat en 4 passos per a la inscripció ràpida per part d'administradors.
- **Interfície d'usuari i flux Frontend (`frontend/`)**:
  - Detecció automàtica dels paràmetres de consulta `deviceToken` i `email` a `LoginView.vue`.
  - Visualització del distintiu de seguretat PrimeVue (`Tag` verd amb icona `pi pi-shield` "Dispositiu autoritzat i verificat") quan s'accedeix mitjançant l'agent.
  - Pre-ompliment automàtic del correu electrònic a l'entrada de text.
  - Emmagatzematge del `deviceToken` al store Pinia (`useAuthStore`) i tramesa al payload de `POST /auth/login`.
  - Neteja del `deviceToken` en tancar sessió (`clearAuth`).
  - Paritat idiomàtica completa en Català, Castellà i Anglès.

## Compliment funcional
- [OK] **Compliment de l'OpenAPI (`contracts/device_auth.openapi.yaml`)**: Endpoints `/devices/verify`, `/admin/devices` (GET/POST) i `/admin/devices/{id}` (DELETE) implementats amb total fidelitat a les signatures, tipus de dades i codis de resposta acordats.
- [OK] **Compliment de l'especificació tècnica (`contracts/device_auth.spec.md`)**:
  - S'utilitza el UUID immutable de la placa base (descartant adreces MAC volàtils).
  - Tokens d'arrencada d'un sol ús amb TTL de 60 segons i consum transaccional atòmic a PostgreSQL.
  - Suport per a desplegament progressiu en dues fases (desplegament suau inicial i mode estricte configurable per variable d'entorn `ENFORCE_TEACHER_DEVICE_AUTH`).
  - Alumnes i administradors preserven l'accés universal sense dependència de l'agent.
- [OK] **Integració amb el mòdul d'autenticació (`backend/internal/auth/`)**: Camp `deviceToken` integrat a `LoginRequest`, validació delegada mitjançant la interfície `DeviceValidator` sense acoblaments circulars.
- [OK] **Agent autònom i multiplataforma (`agent/`)**: Binaris compilats a `agent/bin/` per a Linux (`amd64`), macOS (`universal` darwin arm64 + amd64) i Windows (`amd64.exe`).
- [OK] **Certificats i signatura de codi (`agent/certs/`)**: Certificats i scripts de generació per a PowerShell i OpenSSL presents i documentats per a entorns Windows Defender SmartScreen.
- [OK] **Frontend integrat (`frontend/src/modules/auth/`)**: Integració reactiva a `LoginView.vue`, `store.ts`, `api.ts` i `types.ts`, acompanyada de la suite completa de proves Vitest (`LoginView.spec.ts`, `store.spec.ts`, `api.spec.ts`).

## Qualitat de codi
- [OK] **Compilació i execució de proves**:
  - `go test -v ./internal/device/...` → 14/14 tests passats.
  - `go test -v ./internal/auth/...` → 16/16 tests passats.
  - `go test -v ./...` a `agent/` → 2/2 tests passats.
  - `pnpm test` a `frontend/` → 169/169 tests passats (31 suites).
  - `pnpm build` a `frontend/` → `vue-tsc -b` sense cap discrepància de tipus.
- [OK] **Gestió d'errors i seguretat**:
  - Ús estricte de `shared.AppError` i payloads OpenAPI estàndard.
  - Tokens generats amb entropia criptogràfica (`crypto/rand`).
  - No hi ha consultes SQL concatenades; ús de consultes parametritzades amb transaccions ACID per al consum de tokens.
  - No s'exposen secrets ni credencials al codi font.
- [OK] **Arquitectura neta i mantenibilitat**: Separació estricta en capes `handler` → `service` → `repository` amb injecció de dependències mitjançant interfícies.

## Homogeneïtat
- [OK] **Estructura de paquets**: El paquet `backend/internal/device/` segueix exactament l'estructura canònica dels altres dominis (`auth`, `user`, `quiz`, `match`).
- [OK] **Enrutament**: Rutes registrades de forma homogènia a l'arrel, a `/api` i a `/api/v1` a `main.go`.
- [OK] **Estil i components de Frontend**: Formulari alineat amb les directrius de disseny d'Encertia emprant PrimeVue (`Card`, `InputText`, `Password`, `Tag`, `Button`).
- [OK] **Internacionalització**: Cobertura i paritat idiomàtica a les tres llengües de l'aplicació (`ca`, `es`, `en`).

## Incidències
Cap incidència bloquejant ni important. S'identifiquen únicament dues observacions menors d'optimització futura:

1. **[Menor / Codi] Format d'etiquetes struct a l'agent (`agent/internal/api/client.go`)**:
   - *Descripció*: L'eina `gofmt` suggereix un petit alineament d'espais en blanc a les etiquetes JSON del struct `DeviceSummary`.
   - *Impacte*: Estètic, no afecta la compilació ni el funcionament.
2. **[Menor / UX] Ocultació d'eco a la consola en mode interactiu (`agent/cmd/main.go`)**:
   - *Descripció*: La funció `readPassword()` utilitza un buffer estàndard per llegir la contrasenya de l'administrador sense desactivar l'eco del terminal.
   - *Suggeriment*: Incorporar `golang.org/x/term` amb `term.ReadPassword(int(syscall.Stdin))` per ocultar els caràcters teclejats.
