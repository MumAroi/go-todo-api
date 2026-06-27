package container

import (
	"todo_api/internal/repository"

	"gorm.io/gorm"
)

type Container struct {
	TodoRepo *repository.TodoRepository
}

func NewContainer(db *gorm.DB) *Container {
	return &Container{
		TodoRepo: repository.NewTodoRepository(db),
	}
}
