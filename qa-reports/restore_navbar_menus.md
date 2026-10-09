# QA — Recuperació de Menús a la Navbar (Cursos, Materials, Usuaris)

**Veredicte**: APTE
**Data**: 2026-10-09

## Entorn de proves
- Execució de tests unitaris de frontend: SÍ (`pnpm test` a `frontend/`, 28 fitxers de test, 152 tests superats al 100%).
- Compilació i comprovació de tipus de frontend: SÍ (`pnpm build` i `pnpm type-check` a `frontend/`, sense cap error de TypeScript ni de Vite).
- Verificació de control d'accés per rols (RBAC): SÍ, alineació 100% entre `AppNavbar.vue` i `router/index.ts`.
- Verificació de claus i18n (`ca.json`, `es.json`, `en.json`): SÍ, traduccions completes per a totes les entrades del menú en els tres idiomes suportats.

## Proves funcionals executades
- Execució de proves unitàries a `frontend/src/components/__tests__/AppNavbar.spec.ts`:
  - `renders brand and navigation links for teacher`: Comprova que un professor veu tots els enllaços estàndard i de docència (Jocs & Quizzes, Avaluacions, Grups, Usuaris, Cursos, Materials) i no veu Mètriques. (PASS)
  - `shows metrics link for admin role`: Comprova que l'administrador visualitza Mètriques & Auditoria a més d'Usuaris, Cursos, Materials i la resta d'opcions. (PASS)
  - `hides teacher/admin links for student role`: Comprova que l'alumne veu Jocs & Quizzes, Grups, Cursos i Materials, però no té accés ni visualitza Avaluacions, Usuaris ni Mètriques. (PASS)
- Suite completa de tests (`pnpm test`): 152/152 tests passats.
- Build de producció (`pnpm build`): compilat amb èxit en 2.05s.

## Compliment funcional
- [OK] **Menú Cursos (`/courses`)**: Reincorporat a `AppNavbar.vue` amb la icona `pi pi-book` i traducció `$t('nav.courses')`. Accessible per a tots els usuaris autenticats (estudiants, professors i administradors).
- [OK] **Menú Materials (`/materials`)**: Reincorporat a `AppNavbar.vue` amb la icona `pi pi-folder-open` i traducció `$t('nav.materials')`. Accessible per a tots els usuaris autenticats.
- [OK] **Menú Usuaris (`/users`)**: Reincorporat a `AppNavbar.vue` amb la icona `pi pi-user-plus` i traducció `$t('nav.users')`. Condicionat correctament a `v-if="canAccessTeacherFeatures"` (Admin i Professor).
- [OK] **Regles RBAC i Coherència amb el Router**:
  - `/users`: `router.ts` requereix `roles: ['admin', 'teacher']`, `AppNavbar` restringeix amb `canAccessTeacherFeatures`.
  - `/courses` i `/materials`: `router.ts` requereix `requiresAuth: true`, `AppNavbar` permet accés a qualsevol usuari registrat.
  - Alumne: no veu `/users`, `/evaluations`, `/metrics` ni `/help/teacher-manual`.

## Qualitat de codi
- [OK] **Tipat i Compilació**: `vue-tsc -b` i `vite build` finalitzen de forma completament neta (0 errors).
- [OK] **Estructura i Neteja de Codi**: Respecta les convencions de Vue 3 Composition API (`<script setup lang="ts">`), estils scoped i classes CSS existents.
- [OK] **Traduccions i18n**: Totes les claus de navegació estan degudament declarades a `ca.json`, `es.json` i `en.json`.
- [OK] **Cobertura de Tests**: `AppNavbar.spec.ts` cobreix exhaustivament els tres rols del sistema (`student`, `teacher`, `admin`).

## Homogeneïtat
- [OK] **Patró de Disseny**: Els enllaços reincorporats utilitzen les mateixes classes (`nav-link`), gestió d'estat actiu (`:class="{ active: isRouteActive(...) }"`) i patrons d'iconografia PrimeIcons que la resta d'elements de la barra de navegació.

## Incidències
Cap incidència detectada. La recuperació dels menús compleix plenament els criteris funcionals, de qualitat i de control d'accés.
