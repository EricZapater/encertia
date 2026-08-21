# Especificació Funcional i Tècnica — Mòdul `quiz` (v1)

## 1. Visió General
El mòdul `quiz` gestiona el cicle de vida complet dels jocs/qüestionaris d'Encertia, des de la creació interactiva de preguntes i respostes d'estil Kahoot fins a la publicació, duplicació i emmagatzematge d'imatges a Cloudflare R2 (o fallback local).

---

## 2. Model de Dades (PostgreSQL)

### 2.1 Taula `quizzes`
- `id` (UUID, Primary Key, `gen_random_uuid()`)
- `creator_id` (UUID, Foreign Key a `users.id`, `ON DELETE CASCADE`)
- `title` (VARCHAR(150), NOT NULL)
- `description` (TEXT, NULL)
- `cover_image_url` (TEXT, NULL)
- `status` (VARCHAR(20), NOT NULL, DEFAULT `'draft'`, CHECK `status IN ('draft', 'published', 'archived')`)
- `tags` (TEXT[], NOT NULL, DEFAULT `'{}'`)
- `created_at` (TIMESTAMPTZ, NOT NULL, DEFAULT `NOW()`)
- `updated_at` (TIMESTAMPTZ, NOT NULL, DEFAULT `NOW()`)
- `deleted_at` (TIMESTAMPTZ, NULL, per a Soft Delete)

### 2.2 Taula `quiz_questions`
- `id` (UUID, Primary Key, `gen_random_uuid()`)
- `quiz_id` (UUID, Foreign Key a `quizzes.id`, `ON DELETE CASCADE`)
- `text` (VARCHAR(500), NOT NULL)
- `image_url` (TEXT, NULL)
- `question_type` (VARCHAR(20), NOT NULL, DEFAULT `'single_choice'`, CHECK `question_type IN ('single_choice', 'multiple_choice')`)
- `time_limit_seconds` (INT, NOT NULL, DEFAULT 20, CHECK `time_limit_seconds IN (5, 10, 20, 30, 60, 90, 120)`)
- `order_index` (INT, NOT NULL, DEFAULT 0)
- `created_at` (TIMESTAMPTZ, NOT NULL, DEFAULT `NOW()`)
- `updated_at` (TIMESTAMPTZ, NOT NULL, DEFAULT `NOW()`)

### 2.3 Taula `quiz_answers`
- `id` (UUID, Primary Key, `gen_random_uuid()`)
- `question_id` (UUID, Foreign Key a `quiz_questions.id`, `ON DELETE CASCADE`)
- `text` (VARCHAR(300), NOT NULL)
- `is_correct` (BOOLEAN, NOT NULL, DEFAULT FALSE)
- `order_index` (INT, NOT NULL, DEFAULT 0, CHECK `order_index >= 0 AND order_index <= 5`)
- `created_at` (TIMESTAMPTZ, NOT NULL, DEFAULT `NOW()`)

### 2.4 Índexs
- `CREATE INDEX idx_quizzes_creator_id ON quizzes(creator_id) WHERE deleted_at IS NULL;`
- `CREATE INDEX idx_quizzes_status ON quizzes(status) WHERE deleted_at IS NULL;`
- `CREATE INDEX idx_quizzes_tags ON quizzes USING GIN(tags);`
- `CREATE INDEX idx_quiz_questions_quiz_id ON quiz_questions(quiz_id);`
- `CREATE INDEX idx_quiz_answers_question_id ON quiz_answers(question_id);`

---

## 3. Regles de Negoci i RBAC

1. **Permisos de Creació**:
   - Qualsevol usuari autenticat (`admin`, `teacher`, `student`) pot crear nous qüestionaris.
2. **Privacitat i Propietat**:
   - Cada qüestionari és privat i vinculat al seu creador (`creator_id`).
   - Un usuari només pot consultar, editar, duplicar o eliminar els seus propis qüestionaris.
   - Els usuaris amb rol `admin` tenen accés total per veure, editar i gestionar els qüestionaris de qualsevol usuari.
3. **Validació de Preguntes i Respostes**:
   - Un qüestionari en estat `published` ha de tenir com a mínim 1 pregunta.
   - Cada pregunta ha de tenir entre 2 i 6 respostes (`order_index` 0 a 5).
   - Per a `single_choice`: Exactament 1 resposta ha d'estar marcada com a `is_correct: true`.
   - Per a `multiple_choice`: Com a mínim 1 resposta (i fins a N) pot estar marcada com a `is_correct: true`.
4. **Duplicació de Qüestionaris (`POST /quizzes/{id}/duplicate`)**:
   - Permet duplicar un qüestionari existent creant una nova instància en estat `draft` assignada a l'usuari que fa la petició (`creator_id`).
   - El payload admet:
     - `includeAnswers` (boolean, **per defecte `false`**):
       - Si `includeAnswers: false` (defecte): Només es copien les metadades del qüestionari i els enunciats/configuració de les preguntes (`text`, `image_url`, `question_type`, `time_limit_seconds`, `order_index`), sense copiar cap de les opcions de resposta (`quiz_answers`).
       - Si `includeAnswers: true`: Es copien tant les preguntes com totes les opcions de resposta originals (`quiz_answers`) amb els seus marcadors `is_correct`.
     - `title` (string, opcional): Títol personalitzat per a la còpia (per defecte `"[Còpia] " + títol_original`).
5. **Pujada d'Imatges (Cloudflare R2)**:
   - Formats permesos: PNG, JPG, JPEG, WEBP, GIF.
   - Mida màxima: 5 MB.
   - Servit via URL pública de Cloudflare R2 (o emmagatzematge local configurable en entorns de desenvolupament).

---

## 4. Disseny Frontend (Mòdul `quizzes`)

### 4.1 Vistes
1. `QuizzesListView.vue` (`/quizzes`):
   - Graella de targetes de qüestionaris de l'usuari amb imatge de portada, títol, tags, nombre de preguntes, data i estat.
   - Filtres reactius: cerca de text, selector d'estat (`draft`, `published`, `archived`) i filtre per tags.
   - Botó "Nou Joc" (`/quizzes/new`).
   - Menú contextual a cada targeta: "Editar", "Duplicar", "Canviar estat", "Eliminar".
   - **Modal de Duplicació**: En prémer "Duplicar", apareix un diàleg que permet definir el nou títol i activar l'opció "Copiar també les respostes" (desactivada per defecte, només copia les preguntes).
2. `QuizEditorView.vue` (`/quizzes/:id/edit` o `/quizzes/new`):
   - **Barra Superior**: Títol del quiz, botó de configuració (títol, descripció, tags, imatge portada, estat), botó "Desar", botó "Previsualitzar" i botó "Sortir".
   - **Panell Lateral Esquerre (Llista de Preguntes)**:
     - Miniatures de preguntes numerades (mostrant temps i petit snippet).
     - Botó "+ Afegir Pregunta".
     - Botons de pujar/baixar ordre de pregunta, duplicar i eliminar.
   - **Panell Central (Editor de Pregunta)**:
     - Input per a l'enunciat de la pregunta.
     - Zona per pujar o eliminar la imatge de la pregunta a R2.
     - Controls de paràmetres: selector de tipus (`single_choice` vs `multiple_choice`) i selector de temps (5s, 10s, 20s, 30s, 60s, 90s, 120s).
     - Graella de respostes interactiva (2 a 6 opcions):
       - 6 colors i formes distintives estil Kahoot:
         1. Vermell (▲ Triangle)
         2. Blau (◆ Rombe)
         3. Groc (● Cercle)
         4. Verd (■ Quadrat)
         5. Lila (★ Estrella)
         6. Taronja (⬡ Hexàgon)
       - Checkbox / Radio de resposta correcta.
       - Botons d'afegir/eliminar opció (mínim 2, màxim 6).
3. `QuizPreviewModal.vue`:
   - Modal interactiu per simular les preguntes amb el temporitzador real abans de llançar partides.

---

## 5. Pla de Proves i Criteris d'Acceptació

### 5.1 Backend (Go)
- [ ] Creació de qüestionaris (`POST /quizzes`) amb validació de títol i assignació de `creator_id`.
- [ ] Llistat paginat amb filtres de cerca, estat i tags (`GET /quizzes`).
- [ ] Recuperació de detall complet amb preguntes i respostes ordenades (`GET /quizzes/{id}`).
- [ ] Actualització atòmica en transacció del qüestionari i les seves preguntes/respostes (`PUT /quizzes/{id}`).
- [ ] Protecció RBAC: Un usuari no pot accedir ni modificar qüestionaris d'altres usuaris (retorna 403 o 404), llevat que sigui `admin`.
- [ ] Duplicació de qüestionaris (`POST /quizzes/{id}/duplicate`):
  - [ ] Per defecte (`includeAnswers: false` o payload buit): Duplica només qüestionari i preguntes (sense cap resposta).
  - [ ] Amb `includeAnswers: true`: Duplica qüestionari, preguntes i totes les respostes amb el seu estat `is_correct`.
  - [ ] Amb `title` personalitzat o títol per defecte `[Còpia] ...`.
- [ ] Pujada d'imatges a R2 (`POST /uploads/images`) amb validació de tipus MIME i mida.
- [ ] Soft delete (`DELETE /quizzes/{id}`).

### 5.2 Frontend (Vue 3 / TypeScript)
- [ ] Llistat de qüestionaris amb paginació i filtres.
- [ ] Creador/Editor interactiu amb afegir, reordenar, duplicar i eliminar preguntes.
- [ ] Suport per a 2-6 respostes amb els 6 colors/formes estil Kahoot.
- [ ] Validació abans de desar: com a mínim 1 resposta correcta seleccionada per pregunta.
- [ ] Previsualitzador de preguntes.
