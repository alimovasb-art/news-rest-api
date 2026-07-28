package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {

	fmt.Print("Starting program\n")

	dns := "postgres://postgres:postgres@localhost:5432/news_db?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dns)
	if err != nil {
		log.Fatalf("Failed to connect db: %v\n", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("DB is not responding : %v\n", err)
	}

	fmt.Println("Successfull connection to PostgreSQL (news_db)!")

	DB = pool

}
