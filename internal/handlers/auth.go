package handlers

import (
	"database/sql"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption Failed"})
		return
	}
	var user models.User
	user.Username = input.Username
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	if err := h.DB.QueryRow(query, input.Username, string(hashedPwd)).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	if err := h.DB.QueryRow(query, input.Username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalud Credentials"})
		return
	}
	token, _ := utils.GenerateToken(user.ID)
	c.SetCookie("auth_token", token, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No user ID found"})
		return
	}
	// Type Assertion
	userID, ok := userIDVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal type mismatch"})
		return
	}
	var user models.User
	query := `SELECT id, username, created_at, updated_at FROM users WHERE id = $1`
	if err := h.DB.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	}
	c.JSON(http.StatusOK, user)
}
