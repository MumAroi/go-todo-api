package users

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo_api/internal/shared/container"
	"todo_api/internal/shared/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,strong_password,min=6"`
}

type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"strong_password,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,strong_password,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func CreateUserHandler(c *container.Container) gin.HandlerFunc {
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

		user := &User{
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		userRepo := c.UserRepo.(*UserRepository)
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

func LoginHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errResponse := utils.ParseValidationError(err)
			ctx.JSON(http.StatusBadRequest, errResponse)
			return
		}

		userRepo := c.UserRepo.(*UserRepository)
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
			"exp":     time.Now().Add(c.Config.JWTExpiration).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(c.Config.JWTSecret))

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, LoginResponse{Token: tokenString})
	}
}

func GetUserHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRepo := c.UserRepo.(*UserRepository)
		user, err := userRepo.GetUsers()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func GetUserByIDHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		userRepo := c.UserRepo.(*UserRepository)
		user, err := userRepo.GetUserByID(uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func UpdateUserHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req UpdateUserRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			if err.Error() == "EOF" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body is required"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := c.UserRepo.(*UserRepository)
		_, err = userRepo.GetUserByID(uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		user := make(map[string]any)

		if req.Email != nil {
			user["email"] = *req.Email
		}
		if req.Password != nil {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "fail to hash password " + err.Error()})
				return
			}
			user["password"] = string(hashedPassword)
		}

		if len(user) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		updated, err := userRepo.UpdateUser(uint(id), user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

func DeleteHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		userRepo := c.UserRepo.(*UserRepository)
		if err := userRepo.DeleteUser(uint(id)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}
