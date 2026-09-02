-- Table: materials
CREATE TABLE IF NOT EXISTS materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    material_type VARCHAR(20) NOT NULL CHECK (material_type IN ('document', 'video')),
    file_url TEXT,
    file_name TEXT,
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    page_count INT NOT NULL DEFAULT 0,
    video_url TEXT,
    video_provider VARCHAR(50) CHECK (video_provider IN ('youtube', 'vimeo', 'external') OR video_provider IS NULL),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_materials_teacher_id ON materials(teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_materials_material_type ON materials(material_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_materials_deleted_at ON materials(deleted_at);

-- Table: unit_materials
CREATE TABLE IF NOT EXISTS unit_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id UUID NOT NULL REFERENCES course_units(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unit_materials_unique ON unit_materials (unit_id, material_id);
CREATE INDEX IF NOT EXISTS idx_unit_materials_unit_id ON unit_materials(unit_id);
CREATE INDEX IF NOT EXISTS idx_unit_materials_material_id ON unit_materials(material_id);

-- Table: material_views
CREATE TABLE IF NOT EXISTS material_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_material_views_material_id ON material_views(material_id);
CREATE INDEX IF NOT EXISTS idx_material_views_student_id ON material_views(student_id);
