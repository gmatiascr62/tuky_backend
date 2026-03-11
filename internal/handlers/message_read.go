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

func MarkChatAsRead(c *gin.Context) {
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

	friendID := strings.TrimSpace(c.Param("friendId"))
	if friendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_friend_id",
			"message": "Falta friendId",
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

	repo := repository.NewMessageRepository(pool)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	count, err := repo.MarkAsRead(ctx, user.ID, friendID)
	if err != nil {
		if err == repository.ErrNotFriends {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "not_friends",
				"message": "Solo podés marcar mensajes de amigos",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "mark_read_failed",
			"message": "No se pudieron marcar los mensajes como leídos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"marked_count": count,
		},
	})
}