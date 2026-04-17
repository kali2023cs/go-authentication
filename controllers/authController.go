package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"gin-auth/config"
	"gin-auth/models"
	"gin-auth/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Helper to get OAuth config (lazy loading after .env is loaded)
func getGoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

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
		if existingUser.IsVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or Phone already exists and is verified"})
			return
		}

		existingUser.Name = input.Name
		existingUser.Password = string(hashedPassword)
		config.DB.Save(&existingUser)
		user = existingUser
	} else {
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

	config.DB.Where("email = ? OR phone = ?", user.Email, user.Phone).Delete(&models.OTP{})
	config.DB.Create(&otp)

	if input.SendTo == "phone" {
		fmt.Printf("\n [MOCK SMS] Sending OTP %s to phone %s\n\n", otpCode, user.Phone)
	} else {
		subject := "Your Verification Code"
		body := fmt.Sprintf("Your OTP code is: %s. It expires in 1 minute.", otpCode)
		err := utils.SendEmail(user.Email, subject, body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
			return
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

	config.DB.Model(&models.User{}).Where("email = ? OR phone = ?", input.Identifier, input.Identifier).Update("is_verified", true)
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

	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your account first"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	issueTokens(c, user.ID)
}

func Logout(c *gin.Context) {
	// 1. Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		// 2. Delete from DB (Revocation)
		config.DB.Where("token = ?", refreshToken).Delete(&models.RefreshToken{})
	}

	// 3. Clear cookies
	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token missing"})
		return
	}

	// 1. Validate token signature and expiry
	userID, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 2. Check if token exists in DB (Rotation detection)
	var storedToken models.RefreshToken
	if err := config.DB.Where("token = ? AND user_id = ?", refreshToken, userID).First(&storedToken).Error; err != nil {
		// If token is valid but NOT in DB, it might have been used before (REUSE DETECTION)
		// Better security: Revoke all tokens for this user
		config.DB.Where("user_id = ?", userID).Delete(&models.RefreshToken{})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token reuse detected. Please login again."})
		return
	}

	// 3. Delete the old token (Consume it)
	config.DB.Delete(&storedToken)

	// 4. Issue new pair (Rotation)
	issueTokens(c, userID)
}

func GoogleLogin(c *gin.Context) {
	url := getGoogleConfig().AuthCodeURL("state") // In production, use a random state
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	token, err := getGoogleConfig().Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Find or Create User
	var user models.User
	if err := config.DB.Where("google_id = ? OR email = ?", googleUser.ID, googleUser.Email).First(&user).Error; err != nil {
		user = models.User{
			Name:       googleUser.Name,
			Email:      googleUser.Email,
			GoogleID:   googleUser.ID,
			IsVerified: true, // Google users are pre-verified
		}
		config.DB.Create(&user)
	} else {
		// Update GoogleID if it wasn't set
		if user.GoogleID == "" {
			user.GoogleID = googleUser.ID
			config.DB.Save(&user)
		}
	}

	issueTokens(c, user.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Google Login successful", "user": user.Name})
}

// issueTokens generates tokens, sets cookies, and stores refresh token in DB
func issueTokens(c *gin.Context, userID uint) {
	access, refresh, err := utils.GenerateAccessAndRefreshTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Store Refresh Token in DB
	expiration := time.Now().Add(time.Hour * 24 * 7)
	refreshTokenEntry := models.RefreshToken{
		UserID:    userID,
		Token:     refresh,
		ExpiresAt: expiration,
	}
	config.DB.Create(&refreshTokenEntry)

	// Set HTTP-only Cookies
	// Note: In production, Set Secure: true
	c.SetCookie("access_token", access, int(time.Minute*15/time.Second), "/", "localhost", false, true)
	c.SetCookie("refresh_token", refresh, int(time.Hour*24*7/time.Second), "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated successfully",
		"access_token": access, // Still return access token for convenience, but cookies are the "secure" way
	})
}
