// internal/database/database.go

package database

import (
	"context"
	"log"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // Postgres driver

	"vultisigner/ent"
)

var Client *ent.Client

func Init() {
	// Open a connection to the database
	drv, err := sql.Open("postgres", "your_database_url")
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	Client = ent.NewClient(ent.Driver(drv))
	if err := Client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
