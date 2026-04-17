package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessTokenSecret  = []byte(getEnv("ACCESS_TOKEN_SECRET", "access_secret"))
	refreshTokenSecret = []byte(getEnv("REFRESH_TOKEN_SECRET", "refresh_secret"))
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GenerateToken generates a simple access token (deprecated, use GenerateAccessAndRefreshTokens)
func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // Shorter duration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessTokenSecret)
}

func GenerateAccessAndRefreshTokens(userID uint) (string, string, error) {
	// Access Token (Short-lived: 15 minutes)
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessString, err := accessToken.SignedString(accessTokenSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (Long-lived: 7 days)
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshString, err := refreshToken.SignedString(refreshTokenSecret)
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

func ValidateAccessToken(tokenString string) (uint, error) {
	return validateToken(tokenString, accessTokenSecret)
}

func ValidateRefreshToken(tokenString string) (uint, error) {
	return validateToken(tokenString, refreshTokenSecret)
}

func validateToken(tokenString string, secret []byte) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, jwt.ErrSignatureInvalid
}
