package database

import (
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens SQLite and runs AutoMigrate for Group, Person, and Charge.
func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.Group{}, &models.Person{}, &models.Charge{}); err != nil {
		return nil, err
	}
	return db, nil
}
