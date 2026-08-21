# Especificació Funcional i Tècnica — Mòdul `evaluation` (v1)

## 1. Visió General i Objectiu

El mòdul `evaluation` permet a professors i administradors consultar, analitzar i qualificar els resultats obtinguts pels alumnes en els quizzes jugats mitjançant el mòdul `match`. La nota final per a cada alumne i quiz és assignada per un humà (professor o admin), però el sistema suggereix automàticament una nota calculada a partir de les dades de partida. Els alumnes no tenen accés a aquest mòdul.

**Relació amb mòduls existents:**
- Depèn de `match` (taules `matches`, `match_players`, `match_answers`) per a les dades brutes de resultats.
- Depèn de `quiz` (taules `quizzes`, `quiz_questions`, `quiz_answers`) per als textos de preguntes i respostes.
- Depèn de `user` per als noms i rols dels participants.

---

## 2. Model de Dades (PostgreSQL)

### 2.1 Taula `evaluations`

Registre de la nota (calculada i/o definitiva) d'**un alumne** per a **un quiz concret** (agrega totes les partides d'aquell quiz en que l'alumne ha participat).

```sql
CREATE TABLE IF NOT EXISTS evaluations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id         UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    calculated_grade DECIMAL(4,2) NOT NULL,
    final_grade     DECIMAL(4,2) CHECK (final_grade >= 0 AND final_grade <= 10),
    is_graded       BOOLEAN NOT NULL DEFAULT FALSE,
    graded_by       UUID REFERENCES users(id),
    graded_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_evaluation_quiz_student UNIQUE (quiz_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_evaluations_quiz_id ON evaluations (quiz_id);
CREATE INDEX IF NOT EXISTS idx_evaluations_student_id ON evaluations (student_id);
```

> **Nota**: No hi ha soft-delete propi en `evaluations` perquè és una taula de registres acadèmics. La seva eliminació lògica queda vinculada al soft-delete del quiz o de l'usuari (via `ON DELETE CASCADE` del quiz; el comportament per baixa d'usuari s'ha de decidir en el ticket corresponent, fora de l'abast d'aquest mòdul).

---

## 3. Regles de Negoci

### 3.1 Control d'accés (RBAC)
- **`admin`**: accés complet a totes les avaluacions, sense restricció de propietat.
- **`teacher`**: accés complet, però **només als quizzes que ha creat** (`quizzes.creator_id == user.id`). Si accedeix a un quiz d'un altre professor, rep `403 Forbidden`.
- **`student`**: cap accés. Tots els endpoints retornen `403 Forbidden`.

### 3.2 Creació automàtica del registre d'avaluació
- Quan una partida (`match`) passa a estat `finished`, el backend crea o actualitza un registre `evaluations` per a **cada alumne que ha participat** en aquella partida.
- Si l'alumne ja tenia un registre previ per aquell quiz (havia jugat una partida anterior del mateix quiz), el camp `calculated_grade` es recalcula (vegeu 3.3).
- El camp `final_grade` i `is_graded` **no es toquen** si el professor ja havia assignat una nota.

### 3.3 Fórmula de la nota calculada (`calculated_grade`)
La nota calculada es basa en el resultat de la **darrera partida** en que l'alumne ha participat per a aquell quiz:

```
calculated_grade = (preguntes_encertades_última_partida / total_preguntes_del_quiz) × 10
```

Exemple: alumne que ha jugat 3 partides del mateix quiz de 5 preguntes:
- Partida 1: 3/5 → 6.00
- Partida 2: 4/5 → 8.00
- Partida 3 (última): 5/5 → 10.00
- `calculated_grade` = **10.00**

La nota es trunca a 2 decimals. El rang és sempre `[0.00, 10.00]`.

### 3.4 Qualificació manual (nota definitiva)
- El professor pot assignar o modificar la `final_grade` (escala `0.00 – 10.00`, 2 decimals màxim).
- El sistema pre-omple el camp d'edició amb la `calculated_grade` actual com a suggeriment.
- En assignar, s'estableix `is_graded = true`, `graded_by = user_id`, `graded_at = NOW()`.
- Una nota ja assignada **pot modificar-se** posteriorment; `graded_at` i `graded_by` s'actualitzen.

### 3.5 Estadístiques globals per quiz
Les mètriques agregades es calculen sobre **totes les partides finalitzades** del quiz:

- **Taxa d'encert per pregunta** (`hitRate`): `respostes_correctes / respostes_totals` (0.0–1.0).
- **Temps de resposta mitjà per pregunta** (`avgResponseTimeMs`): mitjana de `match_answers.response_time_ms`.
- **Distribució de respostes** (`answerDistribution`): per a cada opció, recompte i percentatge sobre el total de respostes.
- **Alumnes sense resposta** (`noAnswerCount`): alumnes que van participar en la partida però no van respondre la pregunta concreta.

Totes les mètriques es computen en temps real des de `match_answers` i `match_players` (no es desnormalitzen).

---

## 4. Endpoints de l'API

### 4.1 `GET /evaluations`
Llista els quizzes que tenen almenys una partida finalitzada.

**Resposta:** array de `{ quizId, quizTitle, totalMatches, totalStudents, gradedCount, lastMatchAt }`.

**RBAC:** `admin` veu tots; `teacher` veu només els seus.

---

### 4.2 `GET /evaluations/quizzes/{quizId}`
Vista d'avaluació completa d'un quiz:

- **`stats`**: array de mètriques per pregunta (secció 3.5). Ordre: `order_index` de `quiz_questions`.
- **`students`**: array d'alumnes participants: `studentId`, `studentName`, `matchesCount`, `calculatedGrade`, `finalGrade` (null si no qualificat), `isGraded`.

**Errors:** `403` si `teacher` accedeix a un quiz que no és seu. `404` si el quiz no existeix o no té partides finalitzades.

---

### 4.3 `GET /evaluations/quizzes/{quizId}/students/{studentId}`
Detall d'un alumne per a un quiz:

- Dades del registre: `calculatedGrade`, `finalGrade`, `isGraded`, `gradedBy`, `gradedAt`.
- Array de **partides** participades: `matchId`, `matchDate`, `score`, `totalQuestions`.
- Per a cada partida, array de **respostes**: `questionId`, `questionText`, `selectedAnswerIds`, `correctAnswerIds`, `isCorrect`, `responseTimeMs`. Si l'alumne no va respondre, `selectedAnswerIds: []`, `isCorrect: false`.

**Errors:** `403`, `404` (quiz, alumne no trobat, o alumne sense participació en cap partida d'aquest quiz).

---

### 4.4 `PUT /evaluations/quizzes/{quizId}/students/{studentId}/grade`
Assigna o modifica la nota definitiva.

**Request:** `{ "finalGrade": 7.50 }`

**Validació:** `finalGrade` ∈ `[0, 10]`, màxim 2 decimals.

**Resposta:** `{ evaluationId, calculatedGrade, finalGrade, isGraded, gradedBy, gradedAt }`.

**Errors:** `400` (nota fora de rang o format invàlid), `403`, `404`.

---

## 5. Vistes del Frontend (`frontend/src/modules/evaluations/`)

### 5.1 `views/EvaluationsListView.vue` — ruta: `/evaluations`
Taula de quizzes amb partides finalitzades:
- Columnes: títol del quiz · partides · alumnes · qualificats/total · data última partida · botó "Veure avaluació".
- Accés via menú lateral principal.

### 5.2 `views/QuizEvaluationView.vue` — ruta: `/evaluations/quizzes/:quizId`
Dues seccions:

**A — Estadístiques globals:**
Taula de preguntes: enunciat · taxa d'encert (%) · temps mitjà (s) · distribució d'opcions (nom + %) · sense resposta.

**B — Taula d'alumnes:**
Per alumne: nom · partides jugades · nota calculada · nota definitiva.
- Si `isGraded`: nota en verd + botó "Editar".
- Si no qualificat: nota calculada en gris + botó "Qualificar".
- Clic a la fila navega a `StudentEvaluationView`.

**Accés addicional:** botó "Avaluar" a la pàgina de detall del quiz (`/quizzes/:id`).

### 5.3 `views/StudentEvaluationView.vue` — ruta: `/evaluations/quizzes/:quizId/students/:studentId`
- Capçalera: nom de l'alumne · nota calculada · nota definitiva actual.
- **Formulari de qualificació**: input numèric pre-omplert amb `calculatedGrade`, rang `0–10`, 2 decimals, botó "Desar nota".
- Per a cada partida: capçalera (data, puntuació X/N) + taula de respostes (pregunta · resposta donada · resposta correcta · ✔/✘ · temps).

---

## 6. Dependències i Impacte sobre Mòduls Existents

- **`match` (backend)**: quan un match passa a `finished`, cal crear/actualitzar els registres `evaluations`. Dues opcions vàlides — a decidir pel picacodis backend:
  - (A) El `match` service importa el `evaluation` service directament.
  - (B) Mecanisme d'event/callback intern per mantenir desacoblament. Preferible si la implementació del match ja usa un patró similar.
- **`quiz` (frontend)**: afegir botó "Avaluar" a la vista de detall del quiz, apuntant a `/evaluations/quizzes/:quizId`. Canvi mínim.
- **`quiz` (backend)**: sense canvis d'API.

---

## 7. Pla de Proves

### Backend
- Creació automàtica d'`evaluations` en finalitzar una partida.
- Recàlcul de `calculated_grade` quan l'alumne juga una segona partida del mateix quiz.
- Que `final_grade` **no** es reseteja si ja estava assignada i s'afegeix una nova partida.
- Validació de rang de `final_grade` (< 0 i > 10 retornen `400`).
- RBAC: `student` rep `403` a tots els endpoints; `teacher` rep `403` accedint a un quiz d'un altre professor.

### Frontend
- Tests unitaris del store Pinia (`useEvaluationStore`).
- Tests del component de qualificació: input validat rang 0–10, pre-omplert amb `calculatedGrade`.
- Verificació que el botó "Avaluar" del mòdul `quiz` navega correctament.
