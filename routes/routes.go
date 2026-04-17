package routes

import (
	"gin-auth/controllers"
	"gin-auth/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	// Auth Endpoints
	r.POST("/register", controllers.Register)
	r.POST("/verify-otp", controllers.VerifyOTP)
	r.POST("/login", controllers.Login)
	r.POST("/logout", controllers.Logout)
	r.POST("/refresh", controllers.RefreshToken)

	// Google OAuth Endpoints
	r.GET("/auth/google/login", controllers.GoogleLogin)
	r.GET("/auth/google/callback", controllers.GoogleCallback)

	// Protected Routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(200, gin.H{
				"message": "Protected Route",
				"user_id": userID,
			})
		})
	}
}
