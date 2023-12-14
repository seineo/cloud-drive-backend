package dao

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitMySQLConn opens a connection pool, and it should be shared within the whole app.
func InitMySQLConn(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to init mysqlEventStore connection: %w", err)
	}
	return db, nil
}

// CloseMySQLConn closes the connection to mysqlEventStore, and it can be used in integration tests.
func CloseMySQLConn(db *gorm.DB) error {
	dbInstance, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sqlDB from gorm: %w", err)
	}
	if err := dbInstance.Close(); err != nil {
		return fmt.Errorf("failed to close DB: %w", err)
	}
	return nil
}
