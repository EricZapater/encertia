# Especificació funcional — Mòdul course (Gestió de Cursos, Unitats i Guió de Classe)

**Versió**: 1.0  
**Estat**: En revisió (Pendent de validació humana)  
**Data**: 2026-09-02  

---

## 1. Objectiu del Mòdul

El mòdul `course` proporciona l'estructura organitzativa de la plataforma Encertia. Permet als professors estructurar les seves assignatures en **cursos** i **unitats didàctiques** (on "unitat" i "classe" són el mateix concepte), incriure-hi alumnes, i dissenyar el **guió de classe** (visor seqüencial de blocs) per a la docència en directe.

---

## 2. Èpiques i Històries d'Usuari

### Èpica 1: Gestió de Cursos
- **HU-COURSE-01 (Creació i edició de curs)**: Com a professor o admin, vull poder crear un curs amb títol, descripció, codi, dates d'inici/fi i estat (`draft`, `active`, `archived`) per organitzar la meva docència.
- **HU-COURSE-02 (Llistat i filtrat de cursos)**: Com a usuari, vull veure els cursos on estic inscrit (si sóc alumne) o els que imparteixo/gestiono (si sóc professor/admin).
- **HU-COURSE-03 (Matriculació d'alumnes)**: Com a professor o admin, vull matricular alumnes (individuals o en bloc) a un curs concret i poder-los gestionar o donar de baixa.

### Èpica 2: Unitats Didàctiques ("Classes")
- **HU-COURSE-04 (CRUD d'unitats didàctiques)**: Com a professor, vull crear, editar, reordenar i esborrar unitats didàctiques dins d'un curs (ex. "Tema 1: Introducció", "Tema 2: Algorismes").
- **HU-COURSE-05 (Vinculació N a N amb qüestionaris)**: Com a professor, vull vincular un o més qüestionaris del mòdul `quiz` a una unitat didàctica concreta. Un mateix qüestionari es pot reutilitzar en múltiples unitats o cursos.

### Èpica 3: Guió de Classe (Visor Seqüencial de Blocs)
- **HU-COURSE-06 (Disseny del guió de classe)**: Com a professor, vull construir un guió de blocs ordenats per a una unitat didàctica. Els blocs poden ser:
  1. **Material**: Visualització d'un rang de pàgines d'un PDF/document (ex. diapos 1-15).
  2. **Qüestionari**: Llançament automàtic d'una partida en directe d'un `quiz` associat.
  3. **Pausa / Torn de preguntes**: Aturada lògica explicativa per interacció oberta amb el grup.
- **HU-COURSE-07 (Duplicació de guió)**: Com a professor, vull poder duplicar un guió de classe existent per reutilitzar-lo en una altra unitat o curs.

### Èpica 4: Seguiment de Progrés de l'Alumne
- **HU-COURSE-08 (Progrés d'unitat per alumne)**: Com a alumne/professor, vull veure l'estat de completat de cada unitat didàctica (`pending`, `in_progress`, `completed`).

---

## 3. Normes de Negoci i Restriccions

1. **Restricció d'accés (RBAC)**:
   - `admin`: Accés total a qualsevol curs.
   - `teacher`: Només pot modificar i gestionar els cursos que ha creat (`teacher_id == me.id`).
   - `student`: Només té accés de lectura als cursos on està inscrit (`course_enrollments`).
2. **"Unitat" i "Classe" són el mateix concepte**: No s'introdueix cap jerarquia intermediària. Un curs conté una llista ordenada d'unitats didàctiques.
3. **Relació N a N entre Unitats i Qüestionaris**: La taula `unit_quizzes` gestiona aquesta associació amb un ordre opcional.
4. **Desacoblament del mòdul `material`**: Els blocs de guió de tipus `material` admeten un `materialId` (opcional per quan s'implementi el mòdul `material`) i un `title` + `pdfUrl` per garantir compatibilitat immediata.
5. **Soft-delete**: Els cursos i les unitats fan servir esborrat lògic (`deleted_at`).

---

## 4. Esquema de Dades Relacional (PostgreSQL)

```sql
-- Cursos
CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
    start_date DATE,
    end_date DATE,
    teacher_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Matriculacions d'alumnes
CREATE TABLE course_enrollments (
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (course_id, student_id)
);

-- Unitats Didàctiques
CREATE TABLE course_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Vinculació N:N Unitat <-> Quiz
CREATE TABLE unit_quizzes (
    unit_id UUID NOT NULL REFERENCES course_units(id) ON DELETE CASCADE,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (unit_id, quiz_id)
);

-- Blocs del Guió de Classe
CREATE TABLE script_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id UUID NOT NULL REFERENCES course_units(id) ON DELETE CASCADE,
    block_type VARCHAR(20) NOT NULL CHECK (block_type IN ('material', 'quiz', 'break')),
    order_index INT NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    -- Campos per bloc de material
    material_id UUID,
    pdf_url TEXT,
    start_page INT,
    end_page INT,
    -- Campos per bloc de quiz
    quiz_id UUID REFERENCES quizzes(id) ON DELETE SET NULL,
    -- Campos per bloc de pausa
    duration_minutes INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Progrés de l'alumne per unitat
CREATE TABLE student_unit_progress (
    unit_id UUID NOT NULL REFERENCES course_units(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (unit_id, student_id)
);
```
