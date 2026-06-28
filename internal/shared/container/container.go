package container

import (
	"todo_api/internal/config"
)

type Container struct {
	Config   *config.Config
	TodoRepo any
	UserRepo any
}

func NewContainer(cfg *config.Config, todoRepo any, userRepo any) *Container {
	return &Container{
		Config:   cfg,
		TodoRepo: todoRepo,
		UserRepo: userRepo,
	}
}
