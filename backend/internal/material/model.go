package material

import (
	"time"

	"github.com/google/uuid"
)

type MaterialType string

const (
	TypeDocument MaterialType = "document"
	TypeVideo    MaterialType = "video"
)

type VideoProvider string

const (
	ProviderYouTube  VideoProvider = "youtube"
	ProviderVimeo    VideoProvider = "vimeo"
	ProviderExternal VideoProvider = "external"
)

// Material represents a material entity in the database.
type Material struct {
	ID            uuid.UUID      `json:"id"`
	Title         string         `json:"title"`
	Description   *string        `json:"description"`
	MaterialType  MaterialType   `json:"materialType"`
	FileURL       *string        `json:"fileUrl"`
	FileName      *string        `json:"fileName"`
	FileSizeBytes *int64         `json:"fileSizeBytes"`
	MIMEType      *string        `json:"mimeType"`
	PageCount     int            `json:"pageCount"`
	VideoURL      *string        `json:"videoUrl"`
	VideoProvider *VideoProvider `json:"videoProvider"`
	TeacherID     uuid.UUID      `json:"teacherId"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     *time.Time     `json:"-"`
}

// MaterialResponse matches OpenAPI spec for a material item.
type MaterialResponse struct {
	ID            uuid.UUID      `json:"id"`
	Title         string         `json:"title"`
	Description   *string        `json:"description"`
	MaterialType  MaterialType   `json:"materialType"`
	FileURL       *string        `json:"fileUrl"`
	FileName      *string        `json:"fileName"`
	FileSizeBytes *int64         `json:"fileSizeBytes"`
	MIMEType      *string        `json:"mimeType"`
	PageCount     int            `json:"pageCount"`
	VideoURL      *string        `json:"videoUrl"`
	VideoProvider *VideoProvider `json:"videoProvider"`
	TeacherID     uuid.UUID      `json:"teacherId"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// ToResponse converts Material entity to MaterialResponse DTO.
func (m *Material) ToResponse() MaterialResponse {
	return MaterialResponse{
		ID:            m.ID,
		Title:         m.Title,
		Description:   m.Description,
		MaterialType:  m.MaterialType,
		FileURL:       m.FileURL,
		FileName:      m.FileName,
		FileSizeBytes: m.FileSizeBytes,
		MIMEType:      m.MIMEType,
		PageCount:     m.PageCount,
		VideoURL:      m.VideoURL,
		VideoProvider: m.VideoProvider,
		TeacherID:     m.TeacherID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// MaterialListResponse represents paginated response for materials list.
type MaterialListResponse struct {
	Items      []MaterialResponse `json:"items"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

// CreateMaterialRequest represents the payload to create a new material.
type CreateMaterialRequest struct {
	Title         string         `json:"title"`
	Description   *string        `json:"description"`
	MaterialType  MaterialType   `json:"materialType"`
	FileURL       *string        `json:"fileUrl"`
	FileName      *string        `json:"fileName"`
	FileSizeBytes *int64         `json:"fileSizeBytes"`
	MIMEType      *string        `json:"mimeType"`
	PageCount     *int           `json:"pageCount"`
	VideoURL      *string        `json:"videoUrl"`
	VideoProvider *VideoProvider `json:"videoProvider"`
}

// UpdateMaterialRequest represents payload to update an existing material.
type UpdateMaterialRequest struct {
	Title         *string        `json:"title"`
	Description   *string        `json:"description"`
	FileURL       *string        `json:"fileUrl"`
	FileName      *string        `json:"fileName"`
	FileSizeBytes *int64         `json:"fileSizeBytes"`
	MIMEType      *string        `json:"mimeType"`
	PageCount     *int           `json:"pageCount"`
	VideoURL      *string        `json:"videoUrl"`
	VideoProvider *VideoProvider `json:"videoProvider"`
}

// UploadFileResponse matches OpenAPI spec for uploaded document details.
type UploadFileResponse struct {
	FileURL       string `json:"fileUrl"`
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MIMEType      string `json:"mimeType"`
	PageCount     int    `json:"pageCount"`
}

// StudentViewRecord represents view stats per student for a material.
type StudentViewRecord struct {
	StudentID    uuid.UUID `json:"studentId"`
	StudentName  string    `json:"studentName"`
	StudentEmail string    `json:"studentEmail"`
	ViewCount    int       `json:"viewCount"`
	LastViewedAt time.Time `json:"lastViewedAt"`
}

// MaterialViewsReportResponse matches OpenAPI spec for material views report.
type MaterialViewsReportResponse struct {
	MaterialID          uuid.UUID           `json:"materialId"`
	TotalViews          int                 `json:"totalViews"`
	TotalStudentsViewed int                 `json:"totalStudentsViewed"`
	StudentViews        []StudentViewRecord `json:"studentViews"`
}

// LinkMaterialRequest represents body for linking material to a unit.
type LinkMaterialRequest struct {
	MaterialID uuid.UUID `json:"materialId"`
	OrderIndex *int      `json:"orderIndex"`
}

// MaterialListFilters holds filter parameters for listing materials.
type MaterialListFilters struct {
	Page         int
	PageSize     int
	Search       string
	MaterialType string
}
