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
