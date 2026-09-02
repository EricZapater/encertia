package match

import (
	"context"
	"net/http"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles HTTP and WebSocket endpoints for the Match domain.
type Handler struct {
	service        Service
	tokenValidator shared.TokenValidator
	hub            *Hub
	upgrader       websocket.Upgrader
}

// NewHandler creates a new Handler instance for the Match domain.
func NewHandler(service Service, tokenValidator shared.TokenValidator, hub *Hub) *Handler {
	return &Handler{
		service:        service,
		tokenValidator: tokenValidator,
		hub:            hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allows any frontend origin in dev/prod
			},
		},
	}
}

// RegisterRoutes registers all match endpoints and websocket routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	matchesGroup := rg.Group("/matches")
	{
		matchesGroup.POST("", authMiddleware, h.CreateMatch)
		matchesGroup.GET("/:identifier", h.GetMatchByPin)
		matchesGroup.POST("/:identifier/join", authMiddleware, h.JoinMatch)
		matchesGroup.GET("/:identifier/summary", authMiddleware, h.GetMatchSummary)
	}

	// WebSocket endpoints: supports both /ws/match/:pin and /api/ws/match/:pin
	rg.GET("/ws/match/:pin", h.HandleWebSocket)
}

// CreateMatch handles POST /matches
func (h *Handler) CreateMatch(c *gin.Context) {
	actorID, _, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El camp 'quizId' és obligatori i ha de ser un UUID vàlid.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, err := h.service.CreateMatch(c.Request.Context(), actorID, req.QuizID)
	if err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			shared.RespondWithError(c, appErr)
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	c.JSON(http.StatusCreated, res)
}

// GetMatchByPin handles GET /matches/:pin
func (h *Handler) GetMatchByPin(c *gin.Context) {
	pin := strings.TrimSpace(c.Param("identifier"))
	if len(pin) != 6 {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El PIN ha de ser un codi numèric de 6 dígits.", nil))
		return
	}

	res, err := h.service.GetMatchPublicInfo(c.Request.Context(), pin)
	if err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			shared.RespondWithError(c, appErr)
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	shared.RespondOK(c, res)
}

// JoinMatch handles POST /matches/:pin/join
func (h *Handler) JoinMatch(c *gin.Context) {
	actorID, _, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	pin := strings.TrimSpace(c.Param("identifier"))
	if len(pin) != 6 {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El PIN ha de ser un codi numèric de 6 dígits.", nil))
		return
	}

	var req JoinMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "El sobrenom 'nickname' ha de tenir entre 2 i 30 caràcters.", map[string]interface{}{"raw_error": err.Error()}))
		return
	}

	res, err := h.service.JoinMatch(c.Request.Context(), actorID, pin, req.Nickname)
	if err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			shared.RespondWithError(c, appErr)
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	shared.RespondOK(c, res)
}

// GetMatchSummary handles GET /matches/:id/summary
func (h *Handler) GetMatchSummary(c *gin.Context) {
	actorID, _, appErr := getActorFromContext(c)
	if appErr != nil {
		shared.RespondWithError(c, appErr)
		return
	}

	idParam := strings.TrimSpace(c.Param("identifier"))
	matchID, err := uuid.Parse(idParam)
	if err != nil || matchID == uuid.Nil {
		// If it's a 6-digit PIN, try looking up the match by PIN first
		if len(idParam) == 6 {
			m, findErr := h.service.GetMatchByPIN(c.Request.Context(), idParam)
			if findErr == nil && m != nil {
				matchID = m.ID
			}
		}
		if matchID == uuid.Nil {
			shared.RespondWithError(c, shared.ErrBadRequest(shared.ErrCodeValidation, "L'identificador de partida 'id' ha de ser un UUID vàlid.", nil))
			return
		}
	}

	res, err := h.service.GetMatchSummary(c.Request.Context(), actorID, matchID)
	if err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			shared.RespondWithError(c, appErr)
			return
		}
		shared.RespondWithError(c, shared.ErrInternal(err))
		return
	}

	shared.RespondOK(c, res)
}

// HandleWebSocket handles GET /ws/match/:pin
func (h *Handler) HandleWebSocket(c *gin.Context) {
	pin := strings.TrimSpace(c.Param("pin"))
	if len(pin) != 6 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "El PIN ha de tenir 6 dígits"})
		return
	}

	// 1. Authenticate WebSocket connection via token query param or Authorization header
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = strings.TrimSpace(parts[1])
			}
		}
	}

	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Manca el token d'autenticació."})
		return
	}

	userIDStr, _, _, appErr := h.tokenValidator.ValidateAccessToken(c.Request.Context(), tokenStr)
	if appErr != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token d'autenticació invàlid o caducat."})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Identificador d'usuari invàlid al token."})
		return
	}

	// 2. Fetch match information
	m, err := h.service.GetMatchByPIN(c.Request.Context(), pin)
	if err != nil || m == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "La partida no existeix o ja ha finalitzat."})
		return
	}

	// 3. Determine role (Host vs Player)
	requestedRole := strings.ToLower(strings.TrimSpace(c.Query("role")))
	var isHost bool
	var playerID *uuid.UUID
	nickname := "Host"

	if requestedRole == "host" {
		if m.HostID != userID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Només el creador de la partida pot connectar-se com a moderador (Host)."})
			return
		}
		isHost = true
	} else if requestedRole == "player" {
		player, err := h.service.GetPlayerByMatchAndUser(c.Request.Context(), m.ID, userID)
		if err != nil || player == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Cal unir-se a la partida mitjançant l'endpoint REST abans d'iniciar el WebSocket com a jugador."})
			return
		}
		if player.IsKicked {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Has estat expulsat d'aquesta partida."})
			return
		}
		isHost = false
		playerID = &player.ID
		nickname = player.Nickname
	} else {
		// Default fallback if role is not explicitly specified in URL
		player, err := h.service.GetPlayerByMatchAndUser(c.Request.Context(), m.ID, userID)
		if err == nil && player != nil {
			if player.IsKicked {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Has estat expulsat d'aquesta partida."})
				return
			}
			isHost = false
			playerID = &player.ID
			nickname = player.Nickname
		} else if m.HostID == userID {
			isHost = true
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Cal unir-se a la partida mitjançant l'endpoint REST abans d'iniciar el WebSocket."})
			return
		}
	}

	// 4. Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 5. Register client in Hub room
	room := h.hub.GetOrCreateRoom(pin, m.ID, m.HostID)
	client := &Client{
		Hub:      h.hub,
		Room:     room,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		PlayerID: playerID,
		Nickname: nickname,
		IsHost:   isHost,
	}

	room.RegisterClient(client)

	// Notify connection
	h.service.HandleClientConnect(context.Background(), client)

	// 6. Start pump routines
	go client.WritePump()
	client.ReadPump(
		func(c *Client, msg []byte) {
			h.service.HandleWSMessage(context.Background(), c, msg)
		},
		func(c *Client) {
			room.UnregisterClient(c)
			h.service.HandleClientDisconnect(context.Background(), c)
		},
	)
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
