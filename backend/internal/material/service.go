package material

import (
	"context"
	"errors"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

// Service defines business logic contract for the Material domain.
type Service interface {
	CreateMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, req CreateMaterialRequest) (*MaterialResponse, *shared.AppError)
	GetMaterialByID(ctx context.Context, id uuid.UUID) (*MaterialResponse, *shared.AppError)
	UpdateMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, id uuid.UUID, req UpdateMaterialRequest) (*MaterialResponse, *shared.AppError)
	DeleteMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, id uuid.UUID) *shared.AppError
	ListMaterials(ctx context.Context, actorID uuid.UUID, actorRole string, filters MaterialListFilters) (*MaterialListResponse, *shared.AppError)
	RecordMaterialView(ctx context.Context, actorID uuid.UUID, materialID uuid.UUID) (map[string]interface{}, *shared.AppError)
	GetMaterialViewsReport(ctx context.Context, actorID uuid.UUID, actorRole string, materialID uuid.UUID) (*MaterialViewsReportResponse, *shared.AppError)
	ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]MaterialResponse, *shared.AppError)
	LinkMaterialToUnit(ctx context.Context, actorID uuid.UUID, actorRole string, courseID, unitID uuid.UUID, req LinkMaterialRequest) *shared.AppError
	UnlinkMaterialFromUnit(ctx context.Context, actorID uuid.UUID, actorRole string, courseID, unitID, materialID uuid.UUID) *shared.AppError
}

type service struct {
	repo Repository
}

// NewService creates a new Service instance for the Material domain.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, req CreateMaterialRequest) (*MaterialResponse, *shared.AppError) {
	if actorRole != "admin" && actorRole != "teacher" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Només els professors o administradors poden crear materials.")
	}

	title := strings.TrimSpace(req.Title)
	if len(title) < 2 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del material ha de tenir almenys 2 caràcters.", map[string]interface{}{"field": "title"})
	}

	if req.MaterialType != TypeDocument && req.MaterialType != TypeVideo {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El tipus de material ha de ser 'document' o 'video'.", map[string]interface{}{"field": "materialType"})
	}

	if req.MaterialType == TypeVideo && req.VideoProvider != nil {
		p := *req.VideoProvider
		if p != ProviderYouTube && p != ProviderVimeo && p != ProviderExternal {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El proveïdor de vídeo ha de ser 'youtube', 'vimeo' o 'external'.", map[string]interface{}{"field": "videoProvider"})
		}
	}

	pageCount := 0
	if req.PageCount != nil {
		pageCount = *req.PageCount
	}

	mat := &Material{
		Title:         title,
		Description:   req.Description,
		MaterialType:  req.MaterialType,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSizeBytes: req.FileSizeBytes,
		MIMEType:      req.MIMEType,
		PageCount:     pageCount,
		VideoURL:      req.VideoURL,
		VideoProvider: req.VideoProvider,
		TeacherID:     actorID,
	}

	if err := s.repo.Create(ctx, mat); err != nil {
		return nil, shared.ErrInternal(err)
	}

	resp := mat.ToResponse()
	return &resp, nil
}

func (s *service) GetMaterialByID(ctx context.Context, id uuid.UUID) (*MaterialResponse, *shared.AppError) {
	mat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return nil, shared.ErrInternal(err)
	}

	resp := mat.ToResponse()
	return &resp, nil
}

func (s *service) UpdateMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, id uuid.UUID, req UpdateMaterialRequest) (*MaterialResponse, *shared.AppError) {
	mat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole != "admin" && mat.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per modificar aquest material.")
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if len(t) < 2 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del material ha de tenir almenys 2 caràcters.", map[string]interface{}{"field": "title"})
		}
		mat.Title = t
	}

	if req.Description != nil {
		mat.Description = req.Description
	}
	if req.FileURL != nil {
		mat.FileURL = req.FileURL
	}
	if req.FileName != nil {
		mat.FileName = req.FileName
	}
	if req.FileSizeBytes != nil {
		mat.FileSizeBytes = req.FileSizeBytes
	}
	if req.MIMEType != nil {
		mat.MIMEType = req.MIMEType
	}
	if req.PageCount != nil {
		mat.PageCount = *req.PageCount
	}
	if req.VideoURL != nil {
		mat.VideoURL = req.VideoURL
	}
	if req.VideoProvider != nil {
		p := *req.VideoProvider
		if p != "" && p != ProviderYouTube && p != ProviderVimeo && p != ProviderExternal {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El proveïdor de vídeo ha de ser 'youtube', 'vimeo' o 'external'.", map[string]interface{}{"field": "videoProvider"})
		}
		mat.VideoProvider = req.VideoProvider
	}

	if err := s.repo.Update(ctx, mat); err != nil {
		return nil, shared.ErrInternal(err)
	}

	resp := mat.ToResponse()
	return &resp, nil
}

func (s *service) DeleteMaterial(ctx context.Context, actorID uuid.UUID, actorRole string, id uuid.UUID) *shared.AppError {
	mat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return shared.ErrInternal(err)
	}

	if actorRole != "admin" && mat.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per esborrar aquest material.")
	}

	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) ListMaterials(ctx context.Context, actorID uuid.UUID, actorRole string, filters MaterialListFilters) (*MaterialListResponse, *shared.AppError) {
	if actorRole == "student" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Els alumnes no poden llistar el catàleg de materials.")
	}

	var teacherID *uuid.UUID
	if actorRole == "teacher" {
		teacherID = &actorID
	}

	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 10
	}

	items, total, err := s.repo.List(ctx, filters, teacherID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	responses := make([]MaterialResponse, len(items))
	for i, mat := range items {
		responses[i] = mat.ToResponse()
	}

	totalPages := (total + filters.PageSize - 1) / filters.PageSize
	if totalPages == 0 && total == 0 {
		totalPages = 0
	}

	return &MaterialListResponse{
		Items:      responses,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) RecordMaterialView(ctx context.Context, actorID uuid.UUID, materialID uuid.UUID) (map[string]interface{}, *shared.AppError) {
	_, err := s.repo.GetByID(ctx, materialID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return nil, shared.ErrInternal(err)
	}

	viewCount, err := s.repo.RecordView(ctx, materialID, actorID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return map[string]interface{}{
		"success":   true,
		"viewCount": viewCount,
	}, nil
}

func (s *service) GetMaterialViewsReport(ctx context.Context, actorID uuid.UUID, actorRole string, materialID uuid.UUID) (*MaterialViewsReportResponse, *shared.AppError) {
	mat, err := s.repo.GetByID(ctx, materialID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole != "admin" && mat.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per veure les visualitzacions d'aquest material.")
	}

	report, err := s.repo.GetViewsReport(ctx, materialID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return report, nil
}

func (s *service) ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]MaterialResponse, *shared.AppError) {
	materials, err := s.repo.ListUnitMaterials(ctx, unitID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	responses := make([]MaterialResponse, len(materials))
	for i, m := range materials {
		responses[i] = m.ToResponse()
	}

	return responses, nil
}

func (s *service) LinkMaterialToUnit(ctx context.Context, actorID uuid.UUID, actorRole string, courseID, unitID uuid.UUID, req LinkMaterialRequest) *shared.AppError {
	if actorRole != "admin" {
		teacherID, err := s.repo.GetCourseTeacherID(ctx, courseID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return shared.ErrNotFound("COURSE_NOT_FOUND", "El curs especificat no existeix.")
			}
			return shared.ErrInternal(err)
		}
		if teacherID != actorID {
			return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per vincular materials a aquesta unitat.")
		}
	}

	actualCourseID, err := s.repo.GetCourseIDByUnitID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica especificada no existeix.")
		}
		return shared.ErrInternal(err)
	}
	if actualCourseID != courseID {
		return shared.ErrBadRequest(shared.ErrCodeValidation, "La unitat no pertany al curs especificat.", nil)
	}

	_, err = s.repo.GetByID(ctx, req.MaterialID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return shared.ErrNotFound("MATERIAL_NOT_FOUND", "El material especificat no existeix.")
		}
		return shared.ErrInternal(err)
	}

	orderIdx := 0
	if req.OrderIndex != nil {
		orderIdx = *req.OrderIndex
	}

	if err := s.repo.LinkMaterialToUnit(ctx, unitID, req.MaterialID, orderIdx); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) UnlinkMaterialFromUnit(ctx context.Context, actorID uuid.UUID, actorRole string, courseID, unitID, materialID uuid.UUID) *shared.AppError {
	if actorRole != "admin" {
		teacherID, err := s.repo.GetCourseTeacherID(ctx, courseID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return shared.ErrNotFound("COURSE_NOT_FOUND", "El curs especificat no existeix.")
			}
			return shared.ErrInternal(err)
		}
		if teacherID != actorID {
			return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per desvincular materials d'aquesta unitat.")
		}
	}

	actualCourseID, err := s.repo.GetCourseIDByUnitID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica especificada no existeix.")
		}
		return shared.ErrInternal(err)
	}
	if actualCourseID != courseID {
		return shared.ErrBadRequest(shared.ErrCodeValidation, "La unitat no pertany al curs especificat.", nil)
	}

	if err := s.repo.UnlinkMaterialFromUnit(ctx, unitID, materialID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return shared.ErrNotFound("LINK_NOT_FOUND", "El material no està vinculat a aquesta unitat.")
		}
		return shared.ErrInternal(err)
	}

	return nil
}
