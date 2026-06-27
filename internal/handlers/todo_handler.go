package handlers

import (
	"net/http"
	"todo_api/internal/container"
	"todo_api/internal/models"

	"github.com/gin-gonic/gin"
)

type CreateTodoRequest struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

func CreateTodoHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateTodoRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		todo := &models.Todo{
			Title:     req.Title,
			Completed: req.Completed,
		}

		created, err := c.TodoRepo.CreateTodo(todo)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}
