package shared

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RespondWithJSON writes a JSON response with the given status code and payload.
func RespondWithJSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

// RespondOK writes a 200 OK JSON response.
func RespondOK(c *gin.Context, payload interface{}) {
	c.JSON(http.StatusOK, payload)
}

// RespondCreated writes a 201 Created JSON response.
func RespondCreated(c *gin.Context, payload interface{}) {
	c.JSON(http.StatusCreated, payload)
}

// MessageResponse represents generic message response payload.
type MessageResponse struct {
	Message string `json:"message"`
}
