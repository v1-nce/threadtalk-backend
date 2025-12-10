package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/v1-nce/threadtalk-backend/internal/utils"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie missing"})
			return
		}

		userID, err := utils.ParseToken(tokenString)
		if err != nil {
			c.SetCookie("auth_token", "", -1, "/", "", false, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}