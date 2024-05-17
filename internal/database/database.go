package database

import (
	"fmt"
	"log"
	"vultisigner/config"
	"vultisigner/pkg/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func Init() {
	var err error
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DBName, dbConfig.Password, dbConfig.SSLMode)
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB.AutoMigrate(&models.TransactionPolicy{})
}

func Close() {
	DB.Close()
}
