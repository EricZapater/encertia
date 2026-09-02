# Especificació funcional — Mòdul material (Gestió de Material Didàctic, Documents i Vídeo)

**Versió**: 1.0  
**Estat**: En revisió (Pendent de validació humana)  
**Data**: 2026-09-02  

---

## 1. Objectiu del Mòdul

El mòdul `material` gestiona els recursos didàctics (documents PDF, Word i vídeos incrustats d'allotjament extern com Vimeo/YouTube) que els professors associen a les unitats didàctiques. Inclou la visualització integrada de documents PDF pàgina a pàgina (necessària pel visor de guió de classe de `course`), la possibilitat de reajustar/substituir fitxers sense trencar els enllaços existents, i el registre d'accessos dels alumnes.

---

## 2. Èpiques i Històries d'Usuari

### Èpica 1: Pujada i Gestió de Materials
- **HU-MAT-01 (Pujada de documents)**: Com a professor, vull poder pujar documents (PDF, DOCX) d'un màxim de 50 MB per posar-los a disposició dels alumnes.
- **HU-MAT-02 (Enllaçat de vídeos externs)**: Com a professor, vull poder afegir recursos de vídeo incrustats a partir d'URLs externes (Vimeo, YouTube) amb títol i descripció.
- **HU-MAT-03 (Actualització de document)**: Com a professor, vull poder actualitzar o substituir el fitxer d'un material sense canviar-ne l'ID ni trencar la referència que tenen les unitats didàctiques o els alumnes.

### Èpica 2: Reutilització i Associació amb Unitats
- **HU-MAT-04 (Llistat i reutilització)**: Com a professor, vull veure un repositori dels meus materials i poder associar un mateix material a múltiples unitats didàctiques o cursos (relació N a N).

### Èpica 3: Visualització i Visor Integrat
- **HU-MAT-05 (Visor de PDF integrat)**: Com a alumne/professor, vull poder veure documents PDF directament a l'aplicació, navegar pàgina a pàgina i descarregar el fitxer si ho necessito.
- **HU-MAT-06 (Reproductor de vídeo incrustat)**: Com a alumne, vull poder reproduir vídeos incrustats directament a la interfície del mòdul.

### Èpica 4: Seguretat i Seguiment d'Accés
- **HU-MAT-07 (Registre d'accessos / Mètrics pel professor)**: Com a professor, vull veure un informe d'accessos per material que m'indiqui quins alumnes han consultat o descarregat el recurs i en quines dates.
- **HU-MAT-08 (Control d'accés RBAC)**: Els alumnes només tenen accés de lectura als materials de les unitats dels cursos on estan actualment matriculats.

---

## 3. Normes de Negoci i Restriccions

1. **Formats de fitxer permesos**: `.pdf`, `.docx`, `.doc`, `.pptx`, `.ppt`. Mida màxima: 50 MB.
2. **Emmagatzematge**: AWS SigV4 / Cloudflare R2 per a entorns de producció amb fallback local a `/uploads/materials/`.
3. **Substitució transparent**: En actualitzar un fitxer, es manté l'ID del material (`material_id`), la taula d'associacions `unit_materials` i la taula de guió de classe `script_blocks`.
4. **Seguiment d'accés**: Quan un alumne visualitza o descarrega un material, es crea/actualitza un registre a `material_views`.
5. **Soft-delete**: Esborrat lògic mitjançant `deleted_at`.

---

## 4. Esquema de Dades Relacional (PostgreSQL)

```sql
-- Taula de materials didàctics
CREATE TABLE materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    material_type VARCHAR(20) NOT NULL CHECK (material_type IN ('document', 'video')),
    -- Campos per documents
    file_url TEXT,
    file_name VARCHAR(255),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    page_count INT DEFAULT 0,
    -- Campos per vídeos
    video_url TEXT,
    video_provider VARCHAR(50), -- 'youtube', 'vimeo', 'external'
    -- Propietari i metadades
    teacher_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Associació N:N entre Unitats Didàctiques i Materials
CREATE TABLE unit_materials (
    unit_id UUID NOT NULL REFERENCES course_units(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (unit_id, material_id)
);

-- Registre d'accessos / visualitzacions dels alumnes
CREATE TABLE material_views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    view_count INT NOT NULL DEFAULT 1,
    last_viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (material_id, student_id)
);
```
