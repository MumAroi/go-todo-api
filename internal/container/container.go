package container

import (
	"todo_api/internal/config"
	"todo_api/internal/repository"

	"gorm.io/gorm"
)

type Container struct {
	Config   *config.Config
	TodoRepo *repository.TodoRepository
	UserRepo *repository.UserRepository
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	return &Container{
		Config:   cfg,
		TodoRepo: repository.NewTodoRepository(db),
		UserRepo: repository.NewUserRepository(db),
	}
}
