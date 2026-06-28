package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseUrl string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
}
