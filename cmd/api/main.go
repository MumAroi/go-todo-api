package main

import (
	"log"
	"reflect"
	"strings"
	"todo_api/internal/auth"
	"todo_api/internal/config"
	"todo_api/internal/database"
	"todo_api/internal/shared/middleware"
	"todo_api/internal/shared/validators"
	"todo_api/internal/todos"
	"todo_api/internal/users"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance: ", err)
	}
	defer sqlDB.Close()

	todoRepo := todos.NewTodoRepository(db)
	userRepo := users.NewUserRepository(db)

	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validators.RegisterCustomValidators(v)
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
			if name == "-" {
				return ""
			}
			return name
		})
	}

	router.SetTrustedProxies(nil)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "API Is Running!",
			"status":   "success",
			"database": "connected",
		})
	})

	auth.RegisterRoutes(router, userRepo, cfg)

	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	users.RegisterRoutes(protected.Group("/users"), userRepo)
	todos.RegisterRoutes(protected.Group("/todos"), todoRepo)

	router.Run(":" + cfg.AppPort)
}
