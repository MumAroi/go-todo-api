package todos

import (
	"gorm.io/gorm"
)

type TodoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) CreateTodo(todo *Todo) (*Todo, error) {
	if err := r.db.Create(todo).Error; err != nil {
		return nil, err
	}
	return todo, nil
}

func (r *TodoRepository) GetTodos() ([]Todo, error) {
	var todos []Todo
	if err := r.db.Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (r *TodoRepository) GetTodoByID(id uint) (*Todo, error) {
	var todo Todo
	if err := r.db.Where("id = ?", id).First(&todo).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *TodoRepository) GetTodoByUserIdAndID(id uint, userID string) (*Todo, error) {
	var todo Todo
	if err := r.db.Where("id = ?", id).Where("user_id = ?", userID).First(&todo).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *TodoRepository) UpdateTodo(id uint, updates map[string]any) (*Todo, error) {
	if err := r.db.Model(&Todo{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	var todo Todo
	if err := r.db.First(&todo, id).Error; err != nil {
		return nil, err
	}

	return &todo, nil
}

func (r *TodoRepository) DeleteTodo(id uint) error {
	if err := r.db.Delete(&Todo{}, id).Error; err != nil {
		return err
	}
	return nil
}
