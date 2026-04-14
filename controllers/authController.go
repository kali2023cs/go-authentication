package controllers

import (
	"fmt"
	"net/http"
	"time"

	"gin-auth/config"
	"gin-auth/models"
	"gin-auth/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		SendTo   string `json:"send_to"` // "email" or "phone"
	}
	var user models.User

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input"})
		return
	}

	// Hash Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("email = ? OR phone = ?", input.Email, input.Phone).First(&existingUser).Error; err == nil {
		// User exists, check if verified
		if existingUser.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or Phone already exists and is verified"})
			return
		}

		// If not verified, update their info (in case they changed name/password)
		existingUser.Name = input.Name
		existingUser.Password = string(hashedPassword)
		config.DB.Save(&existingUser)
		user = existingUser // Use existing user for the rest of the flow
	} else {
		// New user, create it
		user = models.User{
			Name:     input.Name,
			Email:    input.Email,
			Phone:    input.Phone,
			Password: string(hashedPassword),
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// Generate OTP
	otpCode := utils.GenerateOTP()
	expiresAt := time.Now().Add(1 * time.Minute)

	otp := models.OTP{
		Email:     user.Email,
		Phone:     user.Phone,
		Code:      otpCode,
		ExpiresAt: expiresAt,
	}

	// Save/Update OTP in DB
	config.DB.Where("email = ? OR phone = ?", user.Email, user.Phone).Delete(&models.OTP{})
	config.DB.Create(&otp)

	// Simulate sending / Real Sending
	if input.SendTo == "phone" {
		fmt.Printf("\n [MOCK SMS] Sending OTP %s to phone %s\n\n", otpCode, user.Phone)
	} else {
		subject := "Your Verification Code"
		body := fmt.Sprintf("Your OTP code is: %s. It expires in 1 minute.", otpCode)
		err := utils.SendEmail(user.Email, subject, body)
		if err != nil {
			fmt.Printf("\n [ERROR] Failed to send email: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
			return
		} else {
			fmt.Printf("\n [SUCCESS] Real Email sent to %s\n\n", user.Email)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully. Please verify within 1 minute."})
}

func VerifyOTP(c *gin.Context) {
	var input struct {
		Identifier string `json:"identifier"` // Email or Phone
		OTP        string `json:"otp"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input"})
		return
	}

	var otp models.OTP
	if err := config.DB.Where("(email = ? OR phone = ?) AND code = ?", input.Identifier, input.Identifier, input.OTP).First(&otp).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP Expired"})
		return
	}

	// Mark user as verified
	config.DB.Model(&models.User{}).Where("email = ? OR phone = ?", input.Identifier, input.Identifier).Update("is_verified", true)

	// Clean up OTP
	config.DB.Delete(&otp)

	c.JSON(http.StatusOK, gin.H{"message": "Account verified successfully"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user models.User

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Input"})
		return
	}

	// Find user
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if verified
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your account first"})
		return
	}

	// Compare password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate Token
	token, _ := utils.GenerateToken(user.ID)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
