package main

import (
	"gin-auth/config"
	"gin-auth/models"
	"gin-auth/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.ConnectDB()

	config.DB.AutoMigrate(&models.User{}, &models.OTP{})

	routes.SetupRoutes(r)

	r.Run(":8080")
}
