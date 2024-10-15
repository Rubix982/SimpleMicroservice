package models

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the PostgreSQL connection
func ConnectDatabase() {
	dsn := "host=postgres-service.default.svc.cluster.local user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatal("Failed to connect to database", err)
	}

	// Auto-migrate the Order model
	if err = DB.AutoMigrate(&Order{}); err != nil {
		logrus.Fatal("Failed to auto-migrate the Order model!", err)
		return
	}
}
