package router

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/v1-nce/threadtalk-backend/internal/handlers"
	"github.com/v1-nce/threadtalk-backend/internal/middleware"
)

func SetUpRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// Apply CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "threadtalk-backend",
		})
	})

	// Rate Limiters
	publicLimit := middleware.NewRateLimiter(5, 10).Middleware()
	authLimit := middleware.NewRateLimiter(1, 3).Middleware()

	authHandler := &handlers.AuthHandler{DB: db}
	forumHandler := &handlers.ForumHandler{DB: db}

	// Public Routes
	r.POST("/auth/signup", authLimit, authHandler.Signup)
	r.POST("/auth/login", authLimit, authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)

	// Public Routes with general rate limiting
	r.GET("/topics", publicLimit, forumHandler.GetTopics)
	r.GET("/topics/:topic_id/posts", publicLimit, forumHandler.GetPosts)
	r.GET("/posts/:post_id", publicLimit, forumHandler.GetPostWithComments)

	// Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.POST("/topics", forumHandler.CreateTopic)
		protected.POST("/posts", forumHandler.CreatePost)
		protected.POST("/comments", forumHandler.CreateComment)
		protected.DELETE("/posts/:post_id", authLimit, forumHandler.DeletePost)
		protected.DELETE("/comments/:comment_id", authLimit, forumHandler.DeleteComment)
	}

	return r
}
