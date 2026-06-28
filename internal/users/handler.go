package users

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"strong_password,min=6"`
}

func GetUserHandler(repo *UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := repo.GetUsers()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func GetUserByIDHandler(repo *UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		user, err := repo.GetUserByID(uint(id))
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

func UpdateUserHandler(repo *UserRepository) gin.HandlerFunc {
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

		_, err = repo.GetUserByID(uint(id))
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

		updated, err := repo.UpdateUser(uint(id), user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

func DeleteUserHandler(repo *UserRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := repo.DeleteUser(uint(id)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}
