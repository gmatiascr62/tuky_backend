package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"tukychat/internal/db"
	"tukychat/internal/middleware"
	"tukychat/internal/models"
	"tukychat/internal/repository"
)

type createMessageBody struct {
	Content string `json:"content"`
}

func CreateMessage(c *gin.Context) {
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

	var body createMessageBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid_body",
			"message": "JSON inválido",
		})
		return
	}

	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "empty_content",
			"message": "El mensaje no puede estar vacío",
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

	msg, err := repo.Create(ctx, user.ID, friendID, body.Content)
	if err != nil {
		if err == repository.ErrNotFriends {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "not_friends",
				"message": "Solo podés enviar mensajes a amigos",
			})
			return
		}
		if err.Error() == "empty_content" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "empty_content",
				"message": "El mensaje no puede estar vacío",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "message_create_failed",
			"message": "No se pudo enviar el mensaje",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    msg,
	})
}

func ListMessages(c *gin.Context) {
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

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid_limit",
				"message": "limit inválido",
			})
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid_offset",
				"message": "offset inválido",
			})
			return
		}
		offset = parsed
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

	items, err := repo.ListConversation(ctx, user.ID, friendID, limit, offset)
	if err != nil {
		if err == repository.ErrNotFriends {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "not_friends",
				"message": "Solo podés ver mensajes con amigos",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "messages_list_failed",
			"message": "No se pudo obtener el historial",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}