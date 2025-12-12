package router

import (
	"database/sql"
	"time"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
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

	authHandler := &handlers.AuthHandler{DB: db}
	forumHandler := &handlers.ForumHandler{DB: db}

	// Public Routes
	r.POST("/auth/signup", authHandler.Signup)
	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/logout", authHandler.Logout)
	r.GET("/topics", forumHandler.GetTopics)
	r.GET("/topics/:topic_id/posts", forumHandler.GetPosts)
	r.GET("/posts/:post_id", forumHandler.GetPostWithComments)

	// Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.POST("/topics", forumHandler.CreateTopic)
		protected.POST("/posts", forumHandler.CreatePost)
		protected.POST("/comments", forumHandler.CreateComment)
	}

	return r
}