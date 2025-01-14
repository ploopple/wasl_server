package db

import (
	"acwj/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "postgresql://postgres:GLbdymbuMZIjgDbhYtDLBDvHLGiLjjdL@roundhouse.proxy.rlwy.net:28906/railway"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	fmt.Println("Database connection established")
}

func Migrate() {
	// Auto-migrate the User model with the custom table name
	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")
}
