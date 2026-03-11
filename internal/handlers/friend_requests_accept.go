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

func AcceptFriendRequest(c *gin.Context) {
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

	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing_request_id",
			"message": "Falta el id de la solicitud",
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

	err = repo.Accept(ctx, requestID, user.ID)
	if err != nil {
		switch err.Error() {
		case "not allowed to accept this request":
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "not_allowed",
				"message": "No podés aceptar esta solicitud",
			})
			return
		case "request already processed":
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "request_already_processed",
				"message": "La solicitud ya fue procesada",
			})
			return
		default:
			if err == repository.ErrAlreadyFriends {
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"error":   "already_friends",
					"message": "Ya son amigos",
				})
				return
			}

			if err == repository.ErrTargetProfileNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   "request_not_found",
					"message": "Solicitud no encontrada",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "friend_request_accept_failed",
				"message": "No se pudo aceptar la solicitud",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status": "accepted",
		},
	})
}