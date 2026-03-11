package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tukychat/internal/db"
	"tukychat/internal/middleware"
	"tukychat/internal/models"
	"tukychat/internal/repository"
)

func Me(c *gin.Context) {
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

	repo := repository.NewProfileRepository(pool)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	profile, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"id":            user.ID,
					"email":         user.Email,
					"role":          user.Role,
					"profile_exists": false,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "profile_fetch_failed",
			"message": "No se pudo obtener el perfil",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":             user.ID,
			"email":          user.Email,
			"role":           user.Role,
			"profile_exists": true,
			"profile":        profile,
		},
	})
}