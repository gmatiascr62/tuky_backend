package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"tukychat/internal/db"
	"tukychat/internal/middleware"
	"tukychat/internal/models"
	"tukychat/internal/repository"
)

type createFriendRequestBody struct {
	UserID string `json:"user_id"`
}

func CreateFriendRequest(c *gin.Context) {
	rawUser, exists := c.Get(middleware.ContextUserKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
			"message": "Usuario no autenticado",
		})
		return
	}

	user, ok := rawUser.(models.AuthUser)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "invalid_auth_context",
			"message": "Contexto de autenticación inválido",
		})
		return
	}

	var body createFriendRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_body",
			"message": "JSON inválido",
		})
		return
	}

	body.UserID = strings.TrimSpace(body.UserID)
	if body.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_user_id",
			"message": "Falta user_id",
		})
		return
	}

	pool, err := db.GetPool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "db_unavailable",
			"message": "No se pudo conectar a la base de datos",
		})
		return
	}

	repo := repository.NewFriendRequestRepository(pool)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	requestID, err := repo.Create(ctx, user.ID, body.UserID)
	if err != nil {
		switch err {
		case repository.ErrCannotFriendSelf:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "cannot_friend_self",
				"message": "No podés enviarte una solicitud a vos mismo",
			})
			return
		case repository.ErrTargetProfileNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "target_profile_not_found",
				"message": "El usuario destino no existe o no tiene perfil",
			})
			return
		case repository.ErrAlreadyFriends:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "already_friends",
				"message": "Ya son amigos",
			})
			return
		case repository.ErrPendingRequestAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "pending_request_exists",
				"message": "Ya existe una solicitud pendiente entre ambos usuarios",
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "friend_request_create_failed",
				"message": "No se pudo crear la solicitud de amistad",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":     requestID,
			"status": "pending",
		},
	})
}