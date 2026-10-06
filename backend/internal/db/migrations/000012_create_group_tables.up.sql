-- Migration up: create groups and group_students tables

CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    academic_year VARCHAR(50) NOT NULL DEFAULT '26-27',
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    teacher_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_groups_academic_year ON groups(academic_year) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_course_id ON groups(course_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_teacher_id ON groups(teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups(deleted_at);

CREATE TABLE IF NOT EXISTS group_students (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_group_students_student_id ON group_students(student_id);
