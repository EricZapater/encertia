# Especificació Tècnica i Funcional: Mòdul `match` (Partides en directe)

## 1. Visió General i Objectiu
El mòdul **`match`** implementa el sistema de partides multijugador interactives en temps real per a Encertia (estil Kahoot). Permet a qualsevol usuari creador (**moderador**) iniciar una partida sobre un qüestionari publicat, generar un codi PIN de 6 dígits i codi QR, projectar la partida i sincronitzar en temps real a tots els **jugadors autenticats** mitjançant WebSockets.

---

## 2. Models de Dades i Esquema PostgreSQL

### 2.1 Taula `matches`
```sql
CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    host_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pin VARCHAR(6) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'lobby' 
        CHECK (status IN ('lobby', 'question_preview', 'question_active', 'question_results', 'leaderboard', 'finished')),
    current_question_index INT NOT NULL DEFAULT 0,
    question_started_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_matches_active_pin ON matches (pin) WHERE status != 'finished' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_matches_host_id ON matches (host_id);
CREATE INDEX IF NOT EXISTS idx_matches_quiz_id ON matches (quiz_id);
```

### 2.2 Taula `match_players`
```sql
CREATE TABLE IF NOT EXISTS match_players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nickname VARCHAR(100) NOT NULL,
    score INT NOT NULL DEFAULT 0,
    is_connected BOOLEAN NOT NULL DEFAULT TRUE,
    is_kicked BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_match_player_user UNIQUE (match_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_match_players_match_id ON match_players (match_id);
CREATE INDEX IF NOT EXISTS idx_match_players_score ON match_players (match_id, score DESC);
```

### 2.3 Taula `match_answers`
```sql
CREATE TABLE IF NOT EXISTS match_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES match_players(id) ON DELETE CASCADE,
    selected_answer_ids UUID[] NOT NULL,
    is_correct BOOLEAN NOT NULL,
    score_awarded INT NOT NULL DEFAULT 0,
    response_time_ms INT NOT NULL DEFAULT 0,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_match_answer_player UNIQUE (match_id, question_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_match_answers_match_question ON match_answers (match_id, question_id);
```

---

## 3. Protocol de Comunicació WebSocket

### 3.1 Connexió i Autenticació
- **URL**: `ws(s)://api.encertia.ericzapater.cat/api/ws/match/:pin?token=JWT_TOKEN`
- Totes les connexions WebSocket requereixen token JWT d'usuari autenticat.
- El servidor identifica si l'usuari és el **moderador** (`host_id == user.id`) o un **jugador** registrat (`user_id == user.id`).

### 3.2 Format dels Missatges (JSON)
```json
{
  "event": "nom_esdeveniment",
  "data": { ... }
}
```

### 3.3 Esdeveniments Client -> Servidor (Accions del Moderador i Jugador)
1. `host:start_match`: El moderador inicia la partida -> passa a la primera pregunta en mode pausa (`question_preview`).
2. `host:start_question_timer`: El moderador acaba la pausa prèvia i activa el temps/opcions (`question_active`).
3. `host:show_results`: El moderador tanca la pregunta i mostra els resultats/respostes correctes (`question_results`).
4. `host:show_leaderboard`: El moderador mostra el rànquing parcial dels jugadors (`leaderboard`).
5. `host:next_question`: El moderador avança a la següent pregunta (`question_preview`) o al podi final si era l'última (`finished`).
6. `host:kick_player`: `{ "playerId": "uuid" }` -> Expulsa un jugador de la sala.
7. `player:submit_answer`: `{ "questionId": "uuid", "answerIds": ["uuid", ...] }` -> Envia la resposta del jugador.

### 3.4 Esdeveniments Servidor -> Client (Broadcast i Notificacions)
1. `match:state`: Sincronització de l'estat actual de la partida, jugadors, índex de pregunta i rol.
2. `match:player_joined`: Notificació de nou jugador a la sala (`{ playerId, nickname, totalPlayers }`).
3. `match:player_left`: Jugador desconnectat (`{ playerId, nickname, totalPlayers }`).
4. `match:player_kicked`: Notificació d'expulsió (`{ playerId }`).
5. `match:question_preview`: Pregunta anunciada (enunciat, imatge, número de pregunta, temps configurat; **respostes encara bloquejades**).
6. `match:question_started`: S'inicia el compte enrere i s'activen les opcions de resposta amb text i colors Kahoot al client.
7. `match:answer_stats`: Actualització en temps real per al moderador de quantes persones han contestat (`{ answeredCount, totalPlayers }`).
8. `match:question_ended`: Es tanca la recepció de respostes i s'envia el resum:
   - Moderador: Recompte per opció de resposta i indicació de les correctes.
   - Jugador: Si la seva resposta ha estat correcta (`isCorrect: true/false`), punts guanyats (`+1` o `+0`) i puntuació total acumulada.
9. `match:leaderboard`: Rànquing parcial ordenat per puntuació (Top jugadors).
10. `match:finished`: Partida finalitzada amb el podi complet (1r, 2n i 3r lloc) i rànquing general.

---

## 4. Màquina d'Estats i Cicle de Vida de la Partida

```
[Lobby] (espera jugadors)
   │
   ▼ host:start_match
[Question Preview] (pausa prèvia: es llegeix l'enunciat)
   │
   ▼ host:start_question_timer
[Question Active] (respostes obertes + compte enrere)
   │
   ▼ temps esgotat o tothom ha respost / host:show_results
[Question Results] (gràfic de barres + revelació de correcta)
   │
   ▼ host:show_leaderboard
[Leaderboard] (rànquing parcial)
   │
   ├── Hi ha més preguntes? ──► host:next_question ──► [Question Preview]
   │
   └── Era l'última pregunta? ──► host:next_question ──► [Finished / Podium]
```

---

## 5. Regles de Negoci
1. **Autenticació Obligatòria**: Tots els participants han d'haver iniciat sessió a Encertia (`admin`, `teacher`, `student`). Si un usuari accedeix a `/play?pin=123456` sense sessió, se'l redirigeix a `/login?redirect=/play?pin=123456`.
2. **Generació de PIN**: PIN numèric aleatori de 6 dígits (ex: `749201`) no utilitzat en cap altra partida activa.
3. **Puntuació**: 1 punt per pregunta encertada (en `single_choice` requereix la resposta correcta; en `multiple_choice` requereix haver seleccionat totes les correctes sense cap incorrecta).
4. **Pausa inicial abans de respondre**: Cada pregunta comença en `question_preview` i només s'obre el temps quan el moderador prem *"Iniciar Temps"*.
5. **Visibilitat de les Respostes**: A la pantalla de l'alumne es mostren els botons grans dels colors i formes Kahoot (▲ Vermell, ◆ Blau, ● Groc, ■ Verd, ★ Lila, ⬡ Taronja) **juntament amb el text de l'opció**.

---

## 6. Vistes del Frontend (`frontend/src/modules/match/`)

1. **`views/PlayerJoinView.vue` (`/play` i `/play?pin=XXXXXX`)**:
   - Formulari interactiu per introduir el PIN de 6 dígits i confirmar o personalitzar el seu Nickname.
2. **`views/PlayerGameView.vue` (`/play/:pin`)**:
   - Pantalla dinàmica de l'alumne segons l'estat:
     - Sala d'espera amb missatge de benvinguda.
     - Pausa de pregunta (enunciat i imatge).
     - Panell de resposta Kahoot amb els 2 a 6 botons de colors/formes i text.
     - Pantalla de feedback immediat (correcte/incorrecte, +1 punt).
     - Rànquing provisional i podi final.
3. **`views/HostLobbyView.vue` (`/matches/:id/lobby` o `/matches/:pin/host`)**:
   - Sala d'espera del moderador amb PIN gran, codi QR interactiu, llista de jugadors connectats, botó d'expulsió i botó "Començar Partida".
4. **`views/HostGameView.vue` (`/matches/:id/game` o `/matches/:pin/host`)**:
   - Panell complet de control de la partida per a projecció:
     - Fase Preview: Enunciat gran, imatge, botó *"Iniciar Temps"*.
     - Fase Activa: Temporitzador per segons, comptador de respostes rebudes en directe, botó *"Tancar i Mostrar Resultats"*.
     - Fase Resultats: Gràfic de barres animat amb la distribució de vots i indicació de la correcta.
     - Fase Rànquing / Podi: Classificació animada i podi dels 3 primers.

---

## 7. Pla de Proves
- **Backend**:
  - Proves de generació i unicitat de PIN.
  - Proves del hub WebSocket (connexió, emissió broadcast, desconnexió).
  - Proves de la màquina d'estats de la partida i càlcul de puntuacions.
- **Frontend**:
  - Tests unitaris del store Pinia (`useMatchStore`).
  - Tests de components per a la vista d'unió (`PlayerJoinView.vue`), pantalla de jugador (`PlayerGameView.vue`) i pantalla de moderador (`HostGameView.vue`).
