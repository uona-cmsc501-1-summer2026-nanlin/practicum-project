package database

import (
	"time"

	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BuiltinCategories are seeded once when the categories table is empty.
var BuiltinCategories = []models.Category{
	{Name: "Food", Icon: "fa-solid fa-utensils", Builtin: true},
	{Name: "Groceries", Icon: "fa-solid fa-cart-shopping", Builtin: true},
	{Name: "Transport", Icon: "fa-solid fa-car", Builtin: true},
	{Name: "Lodging", Icon: "fa-solid fa-bed", Builtin: true},
	{Name: "Entertainment", Icon: "fa-solid fa-film", Builtin: true},
	{Name: "Shopping", Icon: "fa-solid fa-bag-shopping", Builtin: true},
	{Name: "Utilities", Icon: "fa-solid fa-bolt", Builtin: true},
	{Name: "Health", Icon: "fa-solid fa-heart-pulse", Builtin: true},
	{Name: "Travel", Icon: "fa-solid fa-plane", Builtin: true},
	{Name: "Other", Icon: "fa-solid fa-ellipsis", Builtin: true},
}

// Connect opens SQLite and runs AutoMigrate for app models, then seeds builtins.
func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	// SQLite cannot ADD a NOT NULL column without a DEFAULT when rows exist.
	if err := ensureChargeDateColumn(db); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&models.Group{},
		&models.User{},
		&models.GroupMember{},
		&models.Category{},
		&models.Charge{},
	); err != nil {
		return nil, err
	}
	if err := seedBuiltinCategories(db); err != nil {
		return nil, err
	}
	if err := backfillChargeDates(db); err != nil {
		return nil, err
	}
	return db, nil
}

// ensureChargeDateColumn adds charges.date with a DEFAULT before AutoMigrate,
// so existing SQLite rows are valid.
func ensureChargeDateColumn(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Charge{}) {
		return nil
	}
	if db.Migrator().HasColumn(&models.Charge{}, "Date") {
		return nil
	}
	today := time.Now().Format("2006-01-02")
	// SQLite does not allow bound parameters in DEFAULT on ADD COLUMN.
	return db.Exec(
		`ALTER TABLE charges ADD COLUMN date text NOT NULL DEFAULT '` + today + `'`,
	).Error
}

func seedBuiltinCategories(db *gorm.DB) error {
	var n int64
	if err := db.Model(&models.Category{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for i := range BuiltinCategories {
		cat := BuiltinCategories[i]
		if err := db.Create(&cat).Error; err != nil {
			return err
		}
	}
	return nil
}

func backfillChargeDates(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.Charge{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&models.Charge{}, "Date") {
		return nil
	}
	today := time.Now().Format("2006-01-02")
	return db.Exec(
		`UPDATE charges SET date = ? WHERE date IS NULL OR date = '' OR date = '0001-01-01'`,
		today,
	).Error
}
