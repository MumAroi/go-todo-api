package auth

import (
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/shared/utils"
	"todo_api/internal/users"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,strong_password,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,strong_password,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func RegisterHandler(userRepo *users.UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errResponse := utils.ParseValidationError(err)
			ctx.JSON(http.StatusBadRequest, errResponse)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Fail to hash password " + err.Error()})
			return
		}

		user := &users.User{
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		created, err := userRepo.CreateUser(user)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				ctx.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func LoginHandler(userRepo *users.UserRepository, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errResponse := utils.ParseValidationError(err)
			ctx.JSON(http.StatusBadRequest, errResponse)
			return
		}

		user, err := userRepo.GetUserByEmail(req.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(cfg.JWTExpiration).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, LoginResponse{Token: tokenString})
	}
}
