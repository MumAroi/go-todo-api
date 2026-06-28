package todos

import (
	"errors"
	"net/http"
	"strconv"
	"todo_api/internal/shared/container"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateTodoRequest struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

type UpdateTodoRequest struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func CreateTodoHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateTodoRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		todo := &Todo{
			Title:     req.Title,
			Completed: req.Completed,
			UserID:    userID.(string),
		}

		todoRepo := c.TodoRepo.(*TodoRepository)
		created, err := todoRepo.CreateTodo(todo)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, created)
	}
}

func GetTodosHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		todoRepo := c.TodoRepo.(*TodoRepository)
		todos, err := todoRepo.GetTodos()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, todos)
	}
}

func GetTodoByIDHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		todoRepo := c.TodoRepo.(*TodoRepository)
		todo, err := todoRepo.GetTodoByID(uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, todo)
	}
}

func UpdateTodoHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req UpdateTodoRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			if err.Error() == "EOF" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "request body is required"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		todoRepo := c.TodoRepo.(*TodoRepository)
		_, err = todoRepo.GetTodoByUserIdAndID(uint(id), userID.(string))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		todo := make(map[string]any)

		if req.Title != nil {
			todo["title"] = *req.Title
		}
		if req.Completed != nil {
			todo["completed"] = *req.Completed
		}

		if len(todo) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
			return
		}

		updated, err := todoRepo.UpdateTodo(uint(id), todo)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}

func DeleteTodoHandler(c *container.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		todoRepo := c.TodoRepo.(*TodoRepository)
		if err := todoRepo.DeleteTodo(uint(id)); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}
