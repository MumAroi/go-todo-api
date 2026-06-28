package users

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, repo *UserRepository) {
	router.GET("", GetUserHandler(repo))
	router.GET("/:id", GetUserByIDHandler(repo))
	router.PUT("/:id", UpdateUserHandler(repo))
	router.DELETE("/:id", DeleteUserHandler(repo))
}
