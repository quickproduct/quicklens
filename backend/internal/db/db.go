// Package db owns the database connection and schema. GORM (via the pure-Go
// glebarez/modernc driver, CGO-free) owns the *gorm.DB handle and drives the
// schema through AutoMigrate. The underlying *sql.DB is exposed as DB so the
// existing handler/worker queries keep running on the same connection.
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"quicklens/backend/internal/auth"
)

// Gorm is the GORM handle; DB is the shared *sql.DB drawn from the same pool.
var (
	Gorm *gorm.DB
	DB   *sql.DB
)

// Config carries the settings InitDB needs (resolved by the config package).
type Config struct {
	Path          string
	AdminEmail    string
	AdminPassword string
}

// InitDB opens SQLite via GORM/glebarez, applies pragmas, runs AutoMigrate, and
// seeds the default admin user.
func InitDB(cfg Config, logger *zap.Logger) error {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)",
		cfg.Path,
	)

	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	Gorm = gdb

	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("acquire sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	DB = sqlDB

	if err := gdb.AutoMigrate(entityModels()...); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	if err := seedAdmin(cfg, logger); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}
	logger.Info("database initialized")
	return nil
}

func seedAdmin(cfg Config, logger *zap.Logger) error {
	var count int64
	if err := Gorm.Model(&UserEntity{}).Where("email = ?", cfg.AdminEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hashed, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	admin := UserEntity{
		ID:             uuid.New().String(),
		Email:          cfg.AdminEmail,
		HashedPassword: hashed,
		Role:           "admin",
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := Gorm.Create(&admin).Error; err != nil {
		return err
	}
	logger.Info("created default admin user", zap.String("email", cfg.AdminEmail))
	return nil
}
