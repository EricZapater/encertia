DROP INDEX IF EXISTS uq_evaluations_quiz_player;
DROP INDEX IF EXISTS uq_evaluations_quiz_student;

ALTER TABLE evaluations ADD CONSTRAINT uq_evaluation_quiz_student UNIQUE (quiz_id, student_id);

ALTER TABLE evaluations DROP COLUMN IF EXISTS nickname;
ALTER TABLE evaluations DROP COLUMN IF EXISTS player_id;
ALTER TABLE evaluations ALTER COLUMN student_id SET NOT NULL;
