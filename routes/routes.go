package routes

import (
	"gin-auth/controllers"
	"gin-auth/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.POST("/register", controllers.Register)
	r.POST("/verify-otp", controllers.VerifyOTP)
	r.POST("/login", controllers.Login)

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Protected Route"})
		})
	}
}
