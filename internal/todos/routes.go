package todos

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, repo *TodoRepository) {
	router.POST("", CreateTodoHandler(repo))
	router.GET("", GetTodosHandler(repo))
	router.GET("/my", GetMyTodosHandler(repo))
	router.GET("/:id", GetTodoByIDHandler(repo))
	router.PUT("/:id", UpdateTodoHandler(repo))
	router.DELETE("/:id", DeleteTodoHandler(repo))
}
