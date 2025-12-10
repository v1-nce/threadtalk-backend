package router

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/v1-nce/threadtalk-backend/internal/handlers"
	"github.com/v1-nce/threadtalk-backend/internal/middleware"
)

func SetUpRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// Apply CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // Crucial for cookies
		MaxAge:           12 * time.Hour,
	}))

	authHandler := &handlers.AuthHandler{DB: db}

	// Public Routes
	public := r.Group("/auth")
	{
		public.POST("/signup", authHandler.Signup)
		public.POST("/login", authHandler.Login)
		public.POST("/logout", authHandler.Logout)
	}

	// Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("/profile", authHandler.GetProfile)
	}

	return r
}