package auth

import (
	"todo_api/internal/config"
	"todo_api/internal/users"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, userRepo *users.UserRepository, cfg *config.Config) {
	router.POST("/auth/register", RegisterHandler(userRepo))
	router.POST("/auth/login", LoginHandler(userRepo, cfg))
}
