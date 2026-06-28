package main

import (
	"log"
	"reflect"
	"strings"
	"todo_api/internal/config"
	"todo_api/internal/container"
	"todo_api/internal/database"
	"todo_api/internal/handlers"
	"todo_api/internal/middleware"
	"todo_api/internal/validators"

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

	container := container.NewContainer(cfg, db)

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

	router.POST("/auth/register", handlers.CreateUserHandler(container))
	router.POST("/auth/login", handlers.LoginHandler(container))

	protected := router.Group("/todos")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.POST("", handlers.CreateTodoHandler(container))
		protected.GET("", handlers.GetTodosHandler(container))
		protected.GET(":id", handlers.GetTodoByIDHandler(container))
		protected.PUT(":id", handlers.UpdateTodoHandler(container))
		protected.DELETE(":id", handlers.DeleteTodoHandler(container))
	}

	router.Run(":" + cfg.AppPort)
}
