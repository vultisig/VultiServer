package database

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"vultisigner/config"
	"vultisigner/ent"
)

var Client *ent.Client

func Init() {
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DBName, dbConfig.Password, dbConfig.SSLMode)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	Client = client
	// defer client.Close()

	fmt.Println("Connected to database")

	// Run the auto migration tool.
	if err := Client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	fmt.Println("Migration completed")

	// return client
}
