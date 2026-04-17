# Project Setup & Commands

## Initialization
go mod init gin-auth
go mod tidy

## Dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/joho/godotenv
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google

## Run the Application
# Standard run
go run main.go

# Run with hot-reload (if Air is installed)
air