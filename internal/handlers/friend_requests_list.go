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

func ListReceivedFriendRequests(c *gin.Context) {

	rawUser, exists := c.Get(middleware.ContextUserKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
		})
		return
	}

	user := rawUser.(models.AuthUser)

	pool, err := db.GetPool()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	repo := repository.NewFriendRequestRepository(pool)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	requests, err := repo.ListReceived(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}