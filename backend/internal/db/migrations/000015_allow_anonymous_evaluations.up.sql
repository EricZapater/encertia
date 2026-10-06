ALTER TABLE evaluations ALTER COLUMN student_id DROP NOT NULL;
ALTER TABLE evaluations ADD COLUMN IF NOT EXISTS player_id UUID REFERENCES match_players(id) ON DELETE CASCADE;
ALTER TABLE evaluations ADD COLUMN IF NOT EXISTS nickname VARCHAR(100);

ALTER TABLE evaluations DROP CONSTRAINT IF EXISTS uq_evaluation_quiz_student;
DROP INDEX IF EXISTS uq_evaluations_quiz_student;
DROP INDEX IF EXISTS uq_evaluations_quiz_player;

CREATE UNIQUE INDEX IF NOT EXISTS uq_evaluations_quiz_student ON evaluations (quiz_id, student_id) WHERE student_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_evaluations_quiz_player ON evaluations (quiz_id, player_id) WHERE player_id IS NOT NULL;
