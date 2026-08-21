package evaluation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/encertia/backend/internal/shared"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	evals := rg.Group("/evaluations")
	evals.Use(authMiddleware)
	{
		evals.GET("", h.ListEvaluations)
		evals.GET("/quizzes/:quizId", h.GetQuizEvaluation)
		evals.GET("/quizzes/:quizId/students/:studentId", h.GetStudentEvaluation)
		evals.PUT("/quizzes/:quizId/students/:studentId/grade", h.GradeStudent)
	}
}

func (h *Handler) ListEvaluations(c *gin.Context) {
	userID := c.GetString("userId")
	role := c.GetString("userRole")

	summaries, err := h.service.ListEvaluations(userID, role)
	if err != nil {
		if err == ErrUnauthorized {
			shared.RespondWithError(c, shared.ErrForbidden("FORBIDDEN", "Accés no permès a les avaluacions"))
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"evaluations": summaries})
}

func (h *Handler) GetQuizEvaluation(c *gin.Context) {
	quizID := c.Param("quizId")
	userID := c.GetString("userId")
	role := c.GetString("userRole")

	resp, err := h.service.GetQuizEvaluation(quizID, userID, role)
	if err != nil {
		if err == ErrUnauthorized {
			shared.RespondWithError(c, shared.ErrForbidden("FORBIDDEN", "Accés no permès per a aquest quiz"))
			return
		}
		if err.Error() == "quiz not found" {
			shared.RespondWithError(c, shared.ErrNotFound("NOT_FOUND", "Qüestionari no trobat"))
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetStudentEvaluation(c *gin.Context) {
	quizID := c.Param("quizId")
	studentID := c.Param("studentId")
	userID := c.GetString("userId")
	role := c.GetString("userRole")

	resp, err := h.service.GetStudentEvaluation(quizID, studentID, userID, role)
	if err != nil {
		if err == ErrUnauthorized {
			shared.RespondWithError(c, shared.ErrForbidden("FORBIDDEN", "Accés no permès per a aquest alumne"))
			return
		}
		if err.Error() == "student not found" {
			shared.RespondWithError(c, shared.ErrNotFound("NOT_FOUND", "Alumne no trobat"))
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GradeStudent(c *gin.Context) {
	quizID := c.Param("quizId")
	studentID := c.Param("studentId")
	userID := c.GetString("userId")
	role := c.GetString("userRole")

	var req GradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest("BAD_REQUEST", "Format de nota invàlid", nil))
		return
	}

	resp, err := h.service.GradeStudent(quizID, studentID, userID, role, req.FinalGrade)
	if err != nil {
		if err == ErrUnauthorized {
			shared.RespondWithError(c, shared.ErrForbidden("FORBIDDEN", "Accés no permès"))
			return
		}
		if err == ErrInvalidGrade {
			shared.RespondWithError(c, shared.ErrBadRequest("BAD_REQUEST", "La nota ha de ser entre 0.00 i 10.00", nil))
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}
