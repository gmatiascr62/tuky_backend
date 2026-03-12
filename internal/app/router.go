package app

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"tukychat/internal/handlers"
	"tukychat/internal/middleware"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5000",
			"http://127.0.0.1:5000",
			"http://localhost:5500",
			"http://127.0.0.1:5500",
			"https://tuky-front.vercel.app",
		},
		AllowMethods: []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge: 12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", handlers.Health)

		protected := api.Group("/")
		protected.Use(middleware.RequireAuth())
		{
			protected.GET("/me", handlers.Me)
			protected.POST("/profile/setup", handlers.ProfileSetup)
			protected.GET("/users", handlers.ListUsers)

			protected.POST("/friend-requests", handlers.CreateFriendRequest)
			protected.GET("/friend-requests/received", handlers.ListReceivedFriendRequests)
			protected.POST("/friend-requests/:id/accept", handlers.AcceptFriendRequest)

			protected.GET("/friends", handlers.ListFriends)

			protected.POST("/chats/:friendId/messages", handlers.CreateMessage)
			protected.GET("/chats/:friendId/messages", handlers.ListMessages)
			protected.POST("/chats/:friendId/read", handlers.MarkChatAsRead)

			protected.GET("/unread-counts", handlers.GetUnreadCounts)
		}
	}

	return r
}
