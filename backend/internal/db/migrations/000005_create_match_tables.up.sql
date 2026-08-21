-- Create matches table
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

-- Create match_players table
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

-- Create match_answers table
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
