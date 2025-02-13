package database

import (
	"time"

	"github.com/go_/lccxxo/goResourceWatcher/config"
	"github.com/go_/lccxxo/goResourceWatcher/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

// InitDatabase 初始化数据库连接
func InitDatabase() {
	cfg := config.GetConfig()

	// 根据数据库类型选择驱动
	switch cfg.Database.Type {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.Dsn()), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Dsn()), &gorm.Config{})
	default:
		logger.Logger.Panicf("unsupported database type: %s", cfg.Database.Type)
		return
	}

	if err != nil {
		logger.Logger.Panicf("failed to connect to database: %v", err)
		return
	}

	logger.Logger.Infof("connected to %s database", cfg.Database.Type)

	// 设置连接池参数（仅对 MySQL 有意义，SQLite 通常不需要这些设置）
	if cfg.Database.Type == "mysql" {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(2 * time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	}
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if db != nil {
		return db
	}
	logger.Logger.Panic("database not initialized, please call InitDatabase first")
	return nil
}
