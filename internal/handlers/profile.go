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

type setupProfileRequest struct {
	Username string `json:"username"`
}

func ProfileSetup(c *gin.Context) {
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

	var req setupProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_body",
			"message": "JSON inválido",
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 30 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_username",
			"message": "El username debe tener entre 3 y 30 caracteres",
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

	profile, err := repo.Create(ctx, user.ID, req.Username)
	if err != nil {
		switch err {
		case repository.ErrUsernameTaken:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "username_taken",
				"message": "El username ya está en uso",
			})
			return
		case repository.ErrProfileAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "profile_already_exists",
				"message": "El perfil ya existe",
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "profile_create_failed",
				"message": "No se pudo crear el perfil",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    profile,
	})
}