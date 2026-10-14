package device

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the Device domain.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler instance for devices.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all device routes onto the provided router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	rg.POST("/devices/verify", h.VerifyDevice)

	adminDevices := rg.Group("/admin/devices")
	adminDevices.Use(authMiddleware, shared.RequireRole("admin"))
	{
		adminDevices.GET("", h.ListDevices)
		adminDevices.POST("", h.RegisterDevice)
		adminDevices.DELETE("/:id", h.RevokeDevice)
		adminDevices.GET("/download/:platform", h.DownloadAgent)
	}
}

// VerifyDevice handles POST /devices/verify
func (h *Handler) VerifyDevice(c *gin.Context) {
	var req VerifyDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEVICE VERIFY HTTP] Error deserialitzant JSON des de %s: %v", c.ClientIP(), err)
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El cos de la petició no té un format JSON vàlid.",
			map[string]interface{}{"raw_error": err.Error()},
		))
		return
	}

	log.Printf("[DEVICE VERIFY HTTP] Sol·licitud rebuda des de %s per a hash: %s", c.ClientIP(), req.HardwareUUIDHash)
	res, svcErr := h.service.VerifyDevice(c.Request.Context(), req)
	if svcErr != nil {
		log.Printf("[DEVICE VERIFY HTTP ERROR] Error verificant dispositiu (%s): %v", c.ClientIP(), svcErr)
		shared.RespondWithError(c, svcErr)
		return
	}

	log.Printf("[DEVICE VERIFY HTTP SUCCESS] Dispositiu autoritzat: %s (%s). Responent HTTP 200", res.DeviceName, res.TeacherEmail)
	shared.RespondOK(c, res)
}

// ListDevices handles GET /admin/devices
func (h *Handler) ListDevices(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	res, svcErr := h.service.ListDevices(c.Request.Context(), actorRole)
	if svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	shared.RespondOK(c, res)
}

// RegisterDevice handles POST /admin/devices
func (h *Handler) RegisterDevice(c *gin.Context) {
	log.Printf("[DEVICE REGISTER HTTP] Petició de registre d'equip rebuda des de %s (POST /admin/devices)", c.ClientIP())
	actorID, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		log.Printf("[DEVICE REGISTER HTTP WARNING] Autenticació requerida o invàlida: %v", appErr)
		shared.RespondWithError(c, appErr)
		return
	}

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEVICE REGISTER HTTP WARNING] Format JSON invàlid: %v", err)
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El cos de la petició no té un format JSON vàlid.",
			map[string]interface{}{"raw_error": err.Error()},
		))
		return
	}

	log.Printf("[DEVICE REGISTER HTTP] Processant registre per a equip %q (Docent: %s, Hash: %s, AdminID: %s)",
		req.Name, req.TeacherEmail, req.HardwareUUIDHash, actorID)

	res, svcErr := h.service.RegisterDevice(c.Request.Context(), actorID, actorRole, req)
	if svcErr != nil {
		log.Printf("[DEVICE REGISTER HTTP ERROR] Error donant d'alta l'equip: %v", svcErr)
		shared.RespondWithError(c, svcErr)
		return
	}

	log.Printf("[DEVICE REGISTER HTTP SUCCESS] Equip %s (%q) registrat correctament per a %s. Responent HTTP 201 Created",
		res.ID, res.Name, res.TeacherEmail)
	shared.RespondCreated(c, res)
}

// RevokeDevice handles DELETE /admin/devices/:id
func (h *Handler) RevokeDevice(c *gin.Context) {
	_, actorRole, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idStr := c.Param("id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"Format d'identificador de dispositiu (UUID) invàlid.",
			map[string]interface{}{"field": "id"},
		))
		return
	}

	if svcErr := h.service.RevokeDevice(c.Request.Context(), actorRole, deviceID); svcErr != nil {
		shared.RespondWithError(c, svcErr)
		return
	}

	c.Status(http.StatusNoContent)
}

func getActorFromContext(c *gin.Context) (uuid.UUID, string, *shared.AppError) {
	userIDVal, exists := c.Get(shared.CtxKeyUserID)
	if !exists {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "No autenticat.")
	}

	userIDStr, ok := userIDVal.(string)
	if !ok || strings.TrimSpace(userIDStr) == "" {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid.")
	}

	actorID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Identificador d'usuari invàlid.")
	}

	userRoleVal, exists := c.Get(shared.CtxKeyUserRole)
	if !exists {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "Rol d'usuari no especificat al token.")
	}

	actorRole, ok := userRoleVal.(string)
	if !ok || strings.TrimSpace(actorRole) == "" {
		return uuid.Nil, "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Rol d'usuari invàlid.")
	}

	return actorID, actorRole, nil
}

// DownloadAgent handles GET /admin/devices/download/:platform
func (h *Handler) DownloadAgent(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.Param("platform")))

	var fileName string
	var contentType string

	switch platform {
	case "windows", "windows-zip":
		fileName = "encertia-agent-windows.zip"
		contentType = "application/zip"
	case "windows-exe":
		fileName = "encertia-agent-windows.exe"
		contentType = "application/octet-stream"
	case "darwin", "macos":
		fileName = "encertia-agent-darwin"
		contentType = "application/octet-stream"
	case "linux":
		fileName = "encertia-agent-linux"
		contentType = "application/octet-stream"
	case "cert":
		fileName = "encertia-codesign.cer"
		contentType = "application/x-x509-ca-cert"
	default:
		shared.RespondWithError(c, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"Plataforma no suportada. Valors vàlids: windows, windows-zip, windows-exe, darwin, macos, linux, cert.",
			map[string]interface{}{"platform": platform},
		))
		return
	}

	filePath := resolveAgentFilePath(fileName)
	if filePath == "" && (fileName == "encertia-agent-windows.zip") {
		// If zip does not exist yet on disk, build it automatically from exe if present
		exePath := resolveAgentFilePath("encertia-agent-windows.exe")
		if exePath != "" {
			targetZip := filepath.Join(filepath.Dir(exePath), "encertia-agent-windows.zip")
			if err := ensureWindowsZip(exePath, targetZip); err == nil {
				filePath = targetZip
			}
		}
	}

	if filePath == "" {
		shared.RespondWithError(c, shared.ErrNotFound(
			"AGENT_FILE_NOT_FOUND",
			fmt.Sprintf("El fitxer sol·licitat de l'agent (%s) no s'ha trobat al servidor.", fileName),
		))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Header("Content-Type", contentType)
	c.File(filePath)
}

func ensureWindowsZip(exePath string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	// 1. Add exe
	exeFile, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer exeFile.Close()

	exeWriter, err := w.Create("encertia-agent-windows.exe")
	if err != nil {
		return err
	}
	if _, err := io.Copy(exeWriter, exeFile); err != nil {
		return err
	}

	// 2. Add cert if found
	certPath := resolveAgentFilePath("encertia-codesign.cer")
	if certPath != "" && isExistingFile(certPath) {
		if cFile, err := os.Open(certPath); err == nil {
			defer cFile.Close()
			if cWriter, err := w.Create("encertia-codesign.cer"); err == nil {
				_, _ = io.Copy(cWriter, cFile)
			}
		}
	}

	// 3. Add INSTRUCCIONS.txt
	instructions := `ENCERTIA AGENT PER A WINDOWS
============================

Instruccions d'ús:

1. Descomprimeix aquest fitxer ZIP a una carpeta de la teva elecció.

2. Obre una consola de comandes (PowerShell o CMD) a la carpeta i executa:
   .\encertia-agent-windows.exe --register

   Seguiu les instruccions en pantalla per vincular aquest ordinador amb el compte de docent.

3. (Opcional) Si Windows Defender SmartScreen mostra un avís en executar-lo:
   - Fes clic a "Més informació" i seguidament a "Executa de totes maneres".
   - Per afegir el certificat de signatura de codi com a editor de confiança a l'equip i evitar avisos futurs, executa:
     certutil -user -addstore "TrustedPublisher" encertia-codesign.cer
`
	if instWriter, err := w.Create("INSTRUCCIONS.txt"); err == nil {
		_, _ = instWriter.Write([]byte(instructions))
	}

	return nil
}

func resolveAgentFilePath(fileName string) string {
	// 1. Variable d'entorn AGENT_DOWNLOADS_DIR (per defecte /app/downloads)
	downloadsDir := os.Getenv("AGENT_DOWNLOADS_DIR")
	if strings.TrimSpace(downloadsDir) == "" {
		downloadsDir = "/app/downloads"
	}
	candidate := filepath.Join(downloadsDir, fileName)
	if isExistingFile(candidate) {
		return candidate
	}

	// 2. Rutes locals de desenvolupament
	localCandidateDirs := []string{
		"../../agent/bin",
		"../agent/bin",
		"./agent/bin",
		"../../agent/certs",
		"../agent/certs",
		"./agent/certs",
		"agent/bin",
		"agent/certs",
		"../../backend/downloads",
		"../backend/downloads",
		"./downloads",
		"downloads",
	}

	for _, dir := range localCandidateDirs {
		candidate = filepath.Join(dir, fileName)
		if isExistingFile(candidate) {
			return candidate
		}
	}

	return ""
}

func isExistingFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
