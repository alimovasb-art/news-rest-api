package storage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {

	fmt.Print("Starting program\n")

	dns := os.Getenv("DB_CONN")

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
