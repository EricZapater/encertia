# Especificació funcional — Mòdul group (Gestió de Grups d'Alumnes)

**Versió**: 1.0  
**Estat**: En revisió (Pendent de validació humana - Checkpoint 1)  
**Data**: 2026-10-06  

---

## 1. Objectiu del Mòdul

El mòdul `group` proporciona el mestre de **Grups** (ex: "Grup B&F 26-27", "Grup A", etc.) dins de la plataforma Encertia. Permet als professors/administradors organitzar els alumnes en grups per curs i curs acadèmic, facilitant la matriculació col·lectiva i el seguiment de partides o avalucions per grup.

---

## 2. Èpiques i Històries d'Usuari

### Èpica 1: Gestió de Grups (CRUD Mestre)
- **HU-GROUP-01 (Creació de grup)**: Com a professor o admin, vull crear un nou grup introduint el nom (ex: "B&F") i el curs acadèmic (ex: "26-27", amb placeholder "26-27" per defecte), així com opcionalment vincular-lo a un curs/assignatura existent.
- **HU-GROUP-02 (Llistat i filtrat de grups)**: Com a professor o admin, vull veure la llista de grups existents, filtrats per curs acadèmic o per assignatura/curs.
- **HU-GROUP-03 (Edició i esborrat lògic de grup)**: Com a professor o admin, vull modificar les dades d'un grup o donar-lo de baixa (soft-delete) sense perdre l'històric d'activitat dels alumnes.

### Èpica 2: Assignació d'Alumnes al Grup
- **HU-GROUP-04 (Assignació i desassignació d'alumnes)**: Com a professor o admin, vull afegir alumnes (en bloc o individualment) a un grup concret i poder-los veure o desvincular quan calgui.

---

## 3. Normes de Negoci i Restriccions

1. **Camps obligatoris al formulari de creació**:
   - `name` (VARCHAR): Nom del grup (ex: "B&F", "Grup 1A").
   - `academic_year` (VARCHAR): Curs acadèmic (ex: "26-27"). Al formulari el placeholder / proposta per defecte és `"26-27"`.
   - `course_id` (UUID, opcional): Curs/assignatura de la plataforma a la qual s'associa el grup.
2. **Control d'accés (RBAC)**:
   - `admin` / `teacher`: Pot crear, editar, llistar i esborrar grups, així com gestionar els seus integrants.
   - `student`: Només pot veure els grups als quals pertany.
3. **Soft-delete**: Els grups fan servir esborrat lògic (`deleted_at`).
4. **Relació amb Alumnes**: Un alumne pot pertànyer a un o més grups. La taula de vinculació és `group_students`.

---

## 4. Esquema de Dades Relacional (PostgreSQL)

```sql
-- Taula de Grups
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    academic_year VARCHAR(50) NOT NULL DEFAULT '26-27',
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    teacher_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Index per cerques ràpides per curs acadèmic i professor
CREATE INDEX idx_groups_academic_year ON groups(academic_year);
CREATE INDEX idx_groups_course_id ON groups(course_id);
CREATE INDEX idx_groups_teacher_id ON groups(teacher_id);

-- Taula de vinculació Grup <-> Alumnes
CREATE TABLE group_students (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, student_id)
);
```
