DROP INDEX IF EXISTS idx_quiz_answers_question_id;
DROP INDEX IF EXISTS idx_quiz_questions_quiz_id;
DROP INDEX IF EXISTS idx_quizzes_tags;
DROP INDEX IF EXISTS idx_quizzes_status;
DROP INDEX IF EXISTS idx_quizzes_creator_id;

DROP TABLE IF EXISTS quiz_answers;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quizzes;
