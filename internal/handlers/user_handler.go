package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"todo_api/internal/container"
	"todo_api/internal/models"
	"todo_api/internal/utils"

	"github.com/gin-gonic/gin"
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

		user := &models.User{
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		created, err := c.UserRepo.CreateUser(user)
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

func GetUserHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := c.UserRepo.GetUsers()
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

		user, err := c.UserRepo.GetUserByID(uint(id))
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

		_, err = c.UserRepo.GetUserByID(uint(id))
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

		updated, err := c.UserRepo.UpdateUser(uint(id), user)
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

		if err := c.UserRepo.DeleteUser(uint(id)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}
