package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tukychat/internal/db"
	"tukychat/internal/middleware"
	"tukychat/internal/models"
	"tukychat/internal/repository"
)

func GetUnreadCounts(c *gin.Context) {
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

	items, err := repo.GetUnreadCounts(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "unread_counts_failed",
			"message": "No se pudieron obtener los no leídos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}