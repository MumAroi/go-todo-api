package repository

import (
	"todo_api/internal/models"

	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) CreateTodo(todo *models.Todo) (*models.Todo, error) {
	if err := r.db.Create(todo).Error; err != nil {
		return nil, err
	}
	return todo, nil
}

func (r *TodoRepository) GetTodos() ([]models.Todo, error) {
	var todos []models.Todo
	if err := r.db.Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *TodoRepository) GetTodoByID(id uint) (*models.Todo, error) {
	var todo models.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *TodoRepository) UpdateTodo(id uint, updates map[string]any) (*models.Todo, error) {
	if err := r.db.Model(&models.Todo{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	var todo models.Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return nil, err
	}

	return &todo, nil
}

func (r *TodoRepository) DeleteTodo(id uint) error {
	if err := r.db.Delete(&models.Todo{}, id).Error; err != nil {
		return err
	}
	return nil
}
