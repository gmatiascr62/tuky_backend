package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"tukychat/internal/auth"
	"tukychat/internal/config"
	"tukychat/internal/models"
)

const ContextUserKey = "auth_user"

var (
	authClientOnce sync.Once
	authClientInst *auth.Client
)

func getAuthClient() *auth.Client {
	authClientOnce.Do(func() {
		cfg := config.Load()
		authClientInst = auth.NewClient(cfg)
	})
	return authClientInst
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing_authorization_header",
				"message": "Falta el header Authorization",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_authorization_header",
				"message": "El header Authorization debe ser Bearer <token>",
			})
			c.Abort()
			return
		}

		token := strings.TrimSpace(parts[1])

		client := getAuthClient()
		userData, statusCode, err := client.GetUser(token)
		if err != nil {
			c.JSON(statusCode, gin.H{
				"success": false,
				"error":   "auth_verification_failed",
				"message": "No se pudo verificar el token con Supabase",
			})
			c.Abort()
			return
		}

		if userData == nil || userData.ID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_token",
				"message": "Token inválido o expirado",
			})
			c.Abort()
			return
		}

		user := models.AuthUser{
			ID:    userData.ID,
			Email: userData.Email,
			Role:  userData.Role,
		}

		c.Set(ContextUserKey, user)
		c.Next()
	}
}