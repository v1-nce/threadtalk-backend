package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/v1-nce/threadtalk-backend/internal/models"
	"github.com/v1-nce/threadtalk-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB *sql.DB
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var input models.AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if len(input.Username) < 3 || len(input.Username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 3 and 50 characters"})
		return
	}
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ERROR: Failed to hash password for user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	var user models.User
	user.Username = input.Username
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, input.Username, string(hashedPwd)).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if isPgError(err, "23505") {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		log.Printf("ERROR: Failed to create user %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}
	var user models.User
	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, input.Username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("ERROR: Database error during login for username %s: %v", input.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		log.Printf("ERROR: Failed to generate token for user ID %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 3600*24, "/", "", true, true)
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No user ID found"})
		return
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal type mismatch"})
		return
	}
	var user models.User
	query := `SELECT id, username, created_at, updated_at FROM users WHERE id = $1`
	if err := h.DB.QueryRowContext(c.Request.Context(), query, userID).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN: User ID %d not found in database", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("ERROR: Failed to retrieve user ID %d: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
		}
		return
	}
	c.JSON(http.StatusOK, user)
}
