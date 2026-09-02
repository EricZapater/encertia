# Informe d'Auditoria Global QA — Aplicació Encertia

**Veredicte Final Global**: **APTE**  
**Data d'auditoria**: 2026-09-02  
**Mòduls Auditats**: `auth`, `user`, `quiz`, `course`, `match`, `evaluation` (6/6)

---

## 1. Resum Executiu

S'ha realitzat una auditoria i verificació QA integral de TOTA l'aplicació Encertia. Tots els 6 mòduls previstos han estat satisfactòriament implementats, provats i auditats. La suite completa de proves automàtiques del backend en Go i del frontend en Vue 3 / TypeScript s'ha executat amb èxit (100% verda), i l'aplicació s'ha compilat per a producció sense cap avís o error.

---

## 2. Resultats de l'Execució de Suites de Proves

### 2.1 Backend Go (`backend/`)
- **Comanda executada**: `go test -v ./... && go vet ./...`
- **Resultat**: **EXIT CODE 0 (PASS)**
- **Detall de paquets**:
  - `github.com/encertia/backend/internal/auth`: **PASS**
  - `github.com/encertia/backend/internal/user`: **PASS**
  - `github.com/encertia/backend/internal/quiz`: **PASS**
  - `github.com/encertia/backend/internal/course`: **PASS**
  - `github.com/encertia/backend/internal/match`: **PASS**
  - `github.com/encertia/backend/internal/shared`: **PASS**
- **Verificació Go Vet**: Neta, 0 errors o advertències de sintaxi/estàtica.

### 2.2 Frontend Vue 3 / TypeScript (`frontend/`)
- **Comanda executada**: `pnpm run type-check && pnpm run test && pnpm run build`
- **Resultat**: **EXIT CODE 0 (PASS)**
- **Detall**:
  - `vue-tsc -b` (Type-check): **OK** (Sense errors de composició de tipus TypeScript).
  - `vitest run` (Tests unitaris/component): **22 fitxers de test passed, 115 tests passed (115/115)**.
  - `vite build` (Production Build): **OK** (Construcció de paquets finalitzada correctament en `dist/`).

---

## 3. Auditoria per Mòduls

| Mòdul | OpenAPI Contract | Functional Spec | Backend Tests | Frontend Tests | Homogeneïtat | Veredicte |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **auth** | `contracts/auth.openapi.yaml` | `specs/auth.md` | PASS | PASS | OK | **APTE** |
| **user** | `contracts/user.openapi.yaml` | `specs/user.md` | PASS | PASS | OK | **APTE** |
| **quiz** | `contracts/quiz.openapi.yaml` | `specs/quiz.md` | PASS | PASS | OK | **APTE** |
| **course** | `contracts/course.openapi.yaml` | `specs/course.md` | PASS | PASS | OK | **APTE** |
| **match** | `contracts/match.openapi.yaml` | `specs/match.md` | PASS | PASS | OK | **APTE** |
| **evaluation**| `contracts/evaluation.openapi.yaml`| `specs/evaluation.md`| PASS | PASS | OK | **APTE** |

---

## 4. Detall de Compliment Funcional

### 4.1 Autenticació i Sessions (`auth`)
- Par de tokens JWT (Access Token 15 minuts + Refresh Token 7 dies).
- Revocació efectiva de l'Access Token al logout mitjançant taula PostgreSQL `revoked_access_tokens` controlada per `AuthMiddleware`.
- Middleware RBAC per a protecció de rutes segons rol (`admin`, `teacher`, `student`).
- Registre públic d'estudiants (`/auth/register`) i renovació de token (`/auth/refresh`).

### 4.2 Gestió d'Usuaris (`user`)
- Llistat paginat d'usuaris amb cerca i filtratge per rol i estat (`isActive`).
- Control de visibilitat segons el rol: `teacher` només gestiona `student`; `admin` té accés total.
- Alta individual i alta massiva via fitxers batch CSV amb informe de resultats.
- Reseteig de contrasenya administratiu que invalida immediatament les sessions obertes.
- Baixa lògica d'usuaris (`deleted_at = NOW()`) sense borrat físic.

### 4.3 Gestió de Qüestionaris (`quiz`)
- Editor interactiu de qüestionaris amb metadades i etiquetes (`tags`).
- Suport per a preguntes `single_choice` (exactament 1 resposta correcta) i `multiple_choice` (mínim 1).
- Temporitzadors configurables per pregunta (5s a 120s).
- Endpoint de duplicació de qüestionaris amb opció `includeAnswers`.
- Pujada d'imatges amb integració Cloudflare R2 i fallback local en emmagatzematge de desenvolupament.

### 4.4 Cursos, Unitats i Guió de Classe (`course`)
- Creació i llistat de cursos amb matriculació en bloc d'alumnes.
- Unitats didàctiques reordenables amb vinculació N:N reutilitzable a qüestionaris.
- Guió de classe seqüencial per a professors (`ScriptPlayerView`): combina material PDF per pàgines, partides Kahoot en directe i pauses temporitzades.
- Seguiment de l'estat d'aprenentatge per alumne (`pending`, `in_progress`, `completed`).

### 4.5 Partida en Directe (`match`)
- Creació de partides amb codi PIN únic, generació de codi QR i URL d'accés ràpid.
- Connectivitat en temps real via WebSockets (`/ws/match/{matchId}`) per a la sincronització de preguntes, respostes i temporitzador.
- Doble sistema de puntuació: velocitat + encert per al marcador del joc vs encert absolut per a l'avaluació acadèmica.
- Gestió de concurrència amb mutexes (`sync.RWMutex`), reconnexió de jugadors i opció d'expulsió pel professor host.

### 4.6 Avaluació Acadèmica (`evaluation`)
- Generació automàtica d'avaluacions en finalitzar una partida via `RegisterFinishedListener`.
- Estadístiques detallades per qüestionari i pregunta (hit rate, temps mitjà de resposta, respostes no contestades).
- Consulta de la qualificació individual per alumne amb possibilitat d'ajust manual de la nota final (`finalGrade`) pel professor.

---

## 5. Auditoria d'Homogeneïtat i Arquitectura

1. **Nomenclatura dels Rols**: Homologació completa a l'enum `[admin, teacher, student]` a tots els contractes OpenAPI (`contracts/*.openapi.yaml`), especificacions (`specs/*.md`), backend Go (`internal/*/model.go`) i frontend TS/Pinia (`src/modules/*/types.ts`).
2. **Estructura Backend**: Pattern per capes unificat a `backend/internal/<modul>` (`handler.go`, `service.go`, `repository.go`, `model.go`). Utilització estandarditzada del paquet `shared` per a respostes d'error (`ErrorResponse`), middleware i transaccions DB.
3. **Estructura Frontend**: Organització per mòduls a `frontend/src/modules/<moduls>` (`views/`, `components/`, `store.ts`, `api.ts`, `types.ts`). Gestió d'estat centralitzada amb Pinia i crides HTTP unificades via client Axios configurat.
4. **Maneig d'Errors HTTP**: Estandarditzat amb codis de resposta HTTP adequats (400, 401, 403, 404, 409, 500) i cos JSON estructurat.

---

## 6. Veredicte Final

L'aplicació **Encertia** ha superat satisfactòriament totes les auditories funcionals, de qualitat de codi, de compliment OpenAPI i d'homogeneïtat entre mòduls.

**VEREDICTE FINAL GLOBAL**: **APTE PER A FUSIÓ I DESPLEGAMENT**
