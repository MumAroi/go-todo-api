package container

import (
	"todo_api/internal/repository"

	"gorm.io/gorm"
)

type Container struct {
	TodoRepo *repository.TodoRepository
	UserRepo *repository.UserRepository
}

func NewContainer(db *gorm.DB) *Container {
	return &Container{
		TodoRepo: repository.NewTodoRepository(db),
		UserRepo: repository.NewUserRepository(db),
	}
}
