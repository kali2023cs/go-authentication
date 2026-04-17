package middleware

import (
	"net/http"
	"strings"

	"gin-auth/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Try to get token from cookie (Secure practice)
		cookie, err := c.Cookie("access_token")
		if err == nil {
			tokenString = cookie
		} else {
			// 2. Fallback to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// 3. Validate using the utility
		userID, err := utils.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set userID in context for subsequent handlers
		c.Set("user_id", userID)
		c.Next()
	}
}