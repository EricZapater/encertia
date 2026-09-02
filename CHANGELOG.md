# Changelog

Format basat en [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
versionat amb [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.9.0] - 2026-09-02
### Added
- **Contracte d'API i Especificació (`course`)**:
  - Especificació funcional ([`specs/course.md`](file:///Users/eric.zapater/Developer/encertia/specs/course.md)) i contracte OpenAPI 3.0 ([`contracts/course.openapi.yaml`](file:///Users/eric.zapater/Developer/encertia/contracts/course.openapi.yaml)) per al mòdul de gestió de cursos, unitats didàctiques, matriculacions i guió de classe (`course`).
- **Backend (`internal/course`)**:
  - Migració SQL `000008_create_course_tables.up.sql` (`courses`, `course_enrollments`, `course_units`, `unit_quizzes`, `script_blocks`, `student_unit_progress`).
  - Capa de persistència amb SQL pur (`repository.go`), servei de domini amb RBAC (`service.go`) i handlers de Gin HTTP (`handler.go`).
  - Endpoints REST: CRUD de cursos (`/courses`), matriculació en bloc (`/courses/:id/students`), unitats didàctiques i vinculació N:N amb quizzes (`/courses/:id/units`), i disseny/reordenació del guió de classe (`/courses/:id/units/:unitId/script`).
  - Suite de tests unitaris i d'integració a `service_test.go` i `handler_test.go` amb 100% d'èxit.
- **Frontend (`src/modules/courses`)**:
  - Tipus TypeScript del contracte, client d'API Axios (`api.ts`) i store Pinia (`useCourseStore`).
  - Vistes completades:
    - Llistat paginat de cursos amb filtres i cerca ([`CoursesListView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/CoursesListView.vue)).
    - Detall de curs amb organitzador d'unitats i gestor de matriculacions ([`CourseDetailView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/CourseDetailView.vue)).
    - Editor d'unitats i vinculació N:N de quizzes ([`UnitEditorView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/UnitEditorView.vue)).
    - Visor seqüencial / reproductor interactiu de guió de classe per al professor ([`ScriptPlayerView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/courses/views/ScriptPlayerView.vue)).
  - Rutes protegides per rol a Vue Router i activació del menú "Cursos" a `AppNavbar.vue`.
- **QA & Validador**:
  - Informe de QA aprovat amb veredicte **APTE** ([`qa-reports/course.md`](file:///Users/eric.zapater/Developer/encertia/qa-reports/course.md)).

## [0.8.0] - 2026-09-02
### Security
- **Mòdul Auth — Taula de Revocació d'Access Tokens JWT**:
  - Creat la migració SQL `000007_create_revoked_access_tokens.up.sql` per gestionar la taula `revoked_access_tokens` a PostgreSQL.
  - Actualitzat el backend (`internal/auth` i `internal/shared/middleware.go`) per registrar el hash de l'Access Token JWT al fer `/auth/logout` i verificar que no estigui revocat a cada petició protegida (`AuthMiddleware`).
  - Actualitzats el contracte OpenAPI (`contracts/auth.openapi.yaml` v0.2.0) i l'especificació (`specs/auth.md` v1.1) per homologar l'esquema d'autenticació JWT (Access Tokens + Refresh Tokens + Taula de Revocació).
  - Actualitzat l'informe de QA (`qa-reports/auth.md`) amb veredicte **APTE**.

## [0.7.8] - 2026-08-22
### Changed
- **Unificació de Notificacions a Toast en tot el Frontend**:
  - Convertides totes les alertes, missatges d'error i banners de confirmació estàtics (`<Message>`) a notificacions emergents tipus Toast de PrimeVue (`useToast()`).
  - Mòduls actualitzats:
    - **Auth**: `LoginView.vue`, `RegisterView.vue`, `ProfileView.vue`.
    - **Avaluacions**: `EvaluationsListView.vue`, `QuizEvaluationView.vue`, `StudentEvaluationView.vue`.
    - **Usuaris**: `UsersListView.vue`, `UserDetailView.vue`, `UserFormModal.vue`, `ResetPasswordModal.vue`, `BatchImportModal.vue`.
    - **Quizzes**: `QuizzesListView.vue`, `QuizEditorView.vue`, `QuizSettingsModal.vue`, `DuplicateQuizModal.vue`.
    - **Match**: `PlayerJoinView.vue`.

## [0.7.7] - 2026-08-21
### Added
- **Frontend Student Evaluation View (`src/modules/evaluations/views/StudentEvaluationView.vue`)**:
  - Incorporada notificació visual d'èxit via Toast de PrimeVue (`useToast()`) i de banner verd de confirmació (`<Message severity="success">`) en desar correctament una qualificació manual per a un alumne ("Nota de X.XX desada correctament.").
- **Frontend Core Setup (`src/main.ts` & `src/App.vue`)**:
  - Registrat `ToastService` a `main.ts` i integrat el component `<Toast />` global a `App.vue` per habilitar notificacions emergents a tota l'aplicació.

## [0.7.6] - 2026-08-21
### Fixed
- **Backend Evaluation Handler (`internal/evaluation/handler.go`)**:
  - Corregides les claus de lectura de context d'autenticació `userID` i `userRole`. Es feia servir `"userId"` i `"userRole"` en comptes de les constants oficials de `shared.AuthMiddleware` (`shared.CtxKeyUserID` = `"auth_user_id"` i `shared.CtxKeyUserRole` = `"auth_user_role"`). Això provocava que les consultes d'avaluació rebesten `userID = ""` i fallessin amb `INTERNAL_SERVER_ERROR` a PostgreSQL per format de UUID invàlid (`""`).

## [0.7.5] - 2026-08-21
### Fixed
- **Backend Match Service (`internal/match/service.go`)**:
  - Afegida la crida als listeners `MatchFinishedListener` (`s.listeners`) en el moment que la partida passa a estat `StatusFinished` a `handleHostNextQuestion`. Això permet que el servei d'avaluació (`evaluation.Service`) rebi el missatge `OnMatchFinished(matchID)` i calculi/enregistri les notes automàticament a la taula `evaluations`.
- **Frontend Evaluations API (`src/modules/evaluations/api.ts`)**:
  - Corregida l'avaluació de `isMockEnabled` perquè només s'activi si `VITE_USE_MOCKS === 'true'` explícitament. Eliminat el fallback silenciós que retornava `mockData` quan la crida real a l'API d'avaluacions fallava o no trobava dades.

## [0.7.4] - 2026-08-21
### Fixed
- **Backend Match Handler (`internal/match/handler.go`)**:
  - Permesa l'especificació explícita del rol (`?role=player` o `?role=host`) en obrir la connexió WebSocket. Si un usuari `admin` o `teacher` s'uneix com a jugador a una partida que ha creat ell mateix des d'un altre dispositiu (ex. telèfon mòbil), el backend el registra com a **Jugador** (`isHost = false`) en lloc d'unir-lo com a moderador, perllongant així la capacitat d'enviar respostes des del dispositiu mòbil.
- **Frontend Match WS Client & Store (`src/modules/match`)**:
  - `connectAsHost` i `connectAsPlayer` ara envien `role=host` i `role=player` a la URL de connexió WebSocket.

## [0.7.3] - 2026-08-21
### Fixed
- **Frontend Match Store (`src/modules/match/store.ts`)**:
  - Assegurat el reinici de `hasSubmittedAnswer = false`, `mySelectedAnswerIds = []` i `lastAnswerResult = null` quan s'inicia una nova pregunta (`match:question_started` i canvi de `currentQuestionIndex` a `match:state`). Això resol el bloqueig que impedia al segon jugador (o jugadors que havien après en preguntes anteriors) poder contestar en avançar la partida.
- **Backend Match Service (`internal/match/service.go`)**:
  - Afegida notificació d'error explicita `RECORD_ANSWER_FAILED` al client WebSocket en cas de no poder enregistrar la resposta a la base de dades.

## [0.7.2] - 2026-08-21
### Fixed
- **Backend Match Service & Model (`internal/match`)**:
  - Actualitzada l'estructura `MatchStatePayload` a `model.go` i `HandleClientConnect` a `service.go` per incloure les opcions de resposta (`Options`) al camp `CurrentQuestion` quan qualsevol jugador es connecta o es re-sincronitza a la partida (`match:state`).
- **Frontend Match WS Client & Store (`src/modules/match`)**:
  - Actualitzat `wsClient.ts` per processar trampes WebSocket que continguin múltiples missatges JSON separats per salts de línia (`\n`).
  - Protegit `parseQuestionPayload` a `store.ts` per evitar que un payload parcial d'esdeveniment buidi les opcions de resposta prèviament carregades.

## [0.7.1] - 2026-08-21
### Fixed
- **Frontend Match Store (`src/modules/match/store.ts`)**:
  - Afegit la funció de parsatge i normalització `parseQuestionPayload` per als esdeveniments WebSocket `match:question_preview`, `match:question_started` i `match:state`. Corregit l'error on els camps d'enunciat, imatges i opcions venien plans a la resposta del backend i es desestimaven, deixant `currentQuestion` com a `undefined` durant la partida en directe.

## [0.7.0] - 2026-08-21
### Added
- **Frontend UI & Layout Unification**:
  - Nova barra de navegació superior global ([`AppNavbar.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/components/AppNavbar.vue)) accessible en totes les pàgines de la gestió autenticada.
  - Navegació directa entre **Jocs & Quizzes** (`/quizzes`), **Avaluacions** (`/evaluations`), **Usuaris** (`/users`), **Perfil** (`/profile`) i indicador de **Cursos (v2 LMS)**.
  - Filtrat dinàmic de navegació segons el rol de l'usuari (`Admin`, `Professor`, `Alumne`).
  - Ocultació automàtica de la barra de navegació durant les pantalles de joc en directe estil Kahoot (`/play` i `/matches/:id/host`).
  - Tests unitaris a [`AppNavbar.spec.ts`](file:///Users/eric.zapater/Developer/encertia/frontend/src/components/__tests__/AppNavbar.spec.ts).

## [0.6.1] - 2026-08-21
### Fixed
- **Backend (`internal/match` & `internal/evaluation`)**:
  - Corregida la consulta SQL a `match/repository.go` i `evaluation/repository.go` on es feia servir la columna inexistent `u.nickname` i `u.name` en comptes de `TRIM(CONCAT(u.first_name, ' ', u.last_name))`. Aquest error provocava un `QUIZ_NOT_FOUND` al llançar una nova partida (`POST /matches`).

## [0.6.0] - 2026-08-21
### Added
- **Contracte d'API i Especificació**: Especificació funcional ([`contracts/evaluation.spec.md`](file:///Users/eric.zapater/Developer/encertia/contracts/evaluation.spec.md)) i contracte OpenAPI 3.0 ([`contracts/evaluation.openapi.yaml`](file:///Users/eric.zapater/Developer/encertia/contracts/evaluation.openapi.yaml)) per al mòdul d'avaluació posterior pel professor (`evaluation`).
- **Backend (`internal/evaluation`)**:
  - Migració SQL `000006_create_evaluation_tables.up.sql` (taula `evaluations` amb unicitat `(quiz_id, student_id)` i claus foranes `ON DELETE CASCADE`).
  - Lògica de recàlcul automàtic de `calculated_grade` basat en la darrera partida jugada per l'alumne, acotada al rang [0.00, 10.00] i truncada a 2 decimals.
  - Preservació de la nota manual (`final_grade`) sense reset en noves partides.
  - Pattern `MatchFinishedListener` a `match/service.go` per a la notificació automàtica i desacoblada en finalitzar una partida.
  - Endpoints REST: `GET /evaluations`, `GET /evaluations/quizzes/:quizId`, `GET /evaluations/quizzes/:quizId/students/:studentId`, `PUT /evaluations/quizzes/:quizId/students/:studentId/grade`.
  - Protecció RBAC (`admin` tot, `teacher` només els seus quizzes, `student` 403) i suites de unit tests completades.
- **Frontend (`src/modules/evaluations`)**:
  - Tipus TypeScript fidels al contracte OpenAPI, client d'API (`api.ts`), mock fidel (`mockData.ts`) i store Pinia (`useEvaluationStore`).
  - Vista de llista de quizzes avaluables ([`EvaluationsListView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/EvaluationsListView.vue)) a `/evaluations`.
  - Vista d'avaluació global del quiz ([`QuizEvaluationView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/QuizEvaluationView.vue)) amb estadístiques per pregunta i taula d'alumnes amb estat de qualificació.
  - Vista de detall d'alumne ([`StudentEvaluationView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/evaluations/views/StudentEvaluationView.vue)) amb formulari de qualificació manual (0.00-10.00, 2 decimals) i desglossament de respostes per partida.
  - Botó "Avaluar" integrat a l'editor i llistat de quizzes, i rutes protegides per rol (`admin`/`teacher`).

## [0.5.0] - 2026-08-21
### Added
- **Contracte d'API i WebSocket**: Especificació funcional ([`contracts/match.spec.md`](file:///Users/eric.zapater/Developer/encertia/contracts/match.spec.md)) i contracte OpenAPI 3.0 ([`contracts/match.openapi.yaml`](file:///Users/eric.zapater/Developer/encertia/contracts/match.openapi.yaml)) per al mòdul de partides multijugador en temps real (`match`).
- **Backend (`internal/match`)**:
  - Migració SQL `000005_create_match_tables.up.sql` (`matches`, `match_players`, `match_answers`, índexs i unicitat de PIN actiu).
  - Hub de WebSockets concurrent-safe (`sync.RWMutex`) amb suport de pings/pongs de *keep-alive*, diferenciació de rols (Host moderador vs Jugadors) i broadcast de sales per PIN.
  - Màquina d'estats de la partida: `lobby` ➡️ `question_preview` (pausa prèvia de lectura) ➡️ `question_active` (temporitzador regressiu i recepció de respostes) ➡️ `question_results` ➡️ `leaderboard` ➡️ `finished` (podi final).
  - Validació de respostes `single_choice` i `multiple_choice`, puntuació d'1 punt per encert i recompte de respostes en temps real.
  - Endpoints REST: `POST /matches` (creació amb PIN de 6 dígits), `GET /matches/:pin` (estat públic), `POST /matches/:pin/join` (unió d'alumne autenticat), `GET /matches/:id/summary` (resum de partida i podi).
  - Endpoint WebSocket autenticat per JWT: `/ws/match/:pin` i `/api/ws/match/:pin`.
  - Parametrització completa de variables d'entorn (`APP_BASE_URL`, `BASE_URL`, Cloudflare R2).
  - Suite exhaustiva de tests a `hub_test.go`, `service_test.go` i `handler_test.go` amb 100% d'èxit.
- **Frontend (`src/modules/match`)**:
  - Tipus TypeScript, client d'API REST (`api.ts`), client WebSocket resilient (`wsClient.ts`) amb reconexió automàtica i store Pinia (`useMatchStore`).
  - Pantalla d'unió de jugadors ([`PlayerJoinView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/PlayerJoinView.vue)) a `/play` i `/play?pin=...` amb suport de redirecció d'autenticació.
  - Pantalla dinàmica de l'alumne ([`PlayerGameView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/PlayerGameView.vue)) amb els 6 colors i formes Kahoot (▲, ◆, ●, ■, ★, ⬡) **acompanyats del text complet de cada resposta**, retroacció immediata d'encert/error (+1 punt), rànquing i pantalla final.
  - Panell de projecció i control del moderador ([`HostGameView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/HostGameView.vue) / [`HostLobbyView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/match/views/HostLobbyView.vue)):
    - Sala d'espera amb codi PIN gran, **codi QR interactiu natiu**, graella de jugadors connectats i opció d'expulsió.
    - Fase de lectura (Preview) amb botó destacat *"Iniciar Temps"*.
    - Fase de joc activa amb compte enrere i recompte de respostes en temps real.
    - Gràfic de barres animat amb la distribució de vots i indicació de la resposta correcta.
    - Taula de líders parcial i podi 3D animat dels 3 primers classificats (🥇, 🥈, 🥉).
  - Botó d'accés ràpid *"Llançar"* (`pi pi-play`) integrat a [`QuizzesListView.vue`](file:///Users/eric.zapater/Developer/encertia/frontend/src/modules/quizzes/views/QuizzesListView.vue).
  - 18 suites de tests amb **94 tests unitaris i de components superats (100%)**.

## [0.4.0] - 2026-08-21
### Added
- **Contracte d'API**: Especificació funcional ([`contracts/quiz.spec.md`](file:///Users/eric.zapater/Developer/encertia/contracts/quiz.spec.md)) i contracte OpenAPI 3.0 ([`contracts/quiz.openapi.yaml`](file:///Users/eric.zapater/Developer/encertia/contracts/quiz.openapi.yaml)) per al mòdul de jocs/qüestionaris (`quiz`).
- **Backend (`internal/quiz` & `internal/shared/storage`)**:
  - Migració SQL `000003_create_quiz_tables.up.sql` (`quizzes`, `quiz_questions`, `quiz_answers`, índexs i soft delete).
  - Servei d'emmagatzematge d'imatges a Cloudflare R2 mitjançant signatura AWS SigV4 autònoma (amb fallback d'emmagatzematge local a `/app/uploads`).
  - Endpoints REST: `GET /quizzes`, `POST /quizzes`, `GET /quizzes/:id`, `PUT /quizzes/:id`, `DELETE /quizzes/:id`, `POST /quizzes/:id/duplicate`, `POST /uploads/images`.
  - Duplicació personalitzable de qüestionaris: permet copiar només preguntes (per defecte `includeAnswers: false`) o preguntes amb totes les seves respostes (`includeAnswers: true`).
  - Control d'accés RBAC (usuaris només gestionen els seus propis qüestionaris privats, `admin` té accés global).
  - Validacions de negoci d'estat publicat (mínim 1 pregunta; 2 a 6 opcions de resposta amb almenys una correcta).
  - Suite de proves unitàries i d'integració a `service_test.go`, `handler_test.go` i `storage_test.go` amb 100% d'èxit.
- **Frontend (`src/modules/quizzes`)**:
  - Tipus TypeScript, client d'API Axios (`api.ts`) i store Pinia (`useQuizStore`).
  - Vista de llistat de qüestionaris (`QuizzesListView.vue`) amb filtres de cerca, estat i tags, canvi ràpid d'estat i paginador.
  - Modal de duplicació (`DuplicateQuizModal.vue`) amb selecció de títol i opció de copiar respostes (desactivada per defecte).
  - Modal de configuració general (`QuizSettingsModal.vue`) amb pujada d'imatge de portada a Cloudflare R2 i etiquetatge amb PrimeVue Chips.
  - Creador/Editor interactiu (`QuizEditorView.vue`): barra lateral de miniatures de preguntes (afegir, moure amunt/avall, duplicar, eliminar) i panell central d'edició amb 6 colors i formes Kahoot (▲ Vermell, ◆ Blau, ● Groc, ■ Verd, ★ Lila, ⬡ Taronja), selecció de 1 o múltiples correctes, límit de temps (5s-120s) i pujada d'imatge.
  - Simulador de joc (`QuizPreviewModal.vue`) amb compte enrere interactiu per segon, selecció de respostes i pantalla de resultats.
  - Suite de tests unitaris i de components a `__tests__/` amb 69 tests superats.
- **Rutes i Navegació**:
  - Rutes protegides `/quizzes`, `/quizzes/new` i `/quizzes/:id/edit`.
  - Accés directe "Els meus Jocs" a la barra superior i al perfil d'usuari.

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
