package storage

import (
	"context"
	"log"
)

func InitTables() {
	usersTable := `
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY, 
			first_name VARCHAR(25) NOT NULL,
			last_name VARCHAR(25) NOT NULL,
			email VARCHAR(100) NOT NULL UNIQUE CHECK (email ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
			password VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,	
			updated_at	TIMESTAMP,
			deleted_at TIMESTAMP
		)
	`
	_, err := DB.Exec(context.Background(), usersTable)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	log.Printf("Table users successfully created/checked!")

	newsTable := `
		CREATE TABLE IF NOT EXISTS news(
			id SERIAL PRIMARY KEY,
			title VARCHAR(50) NOT NULL,
			short_description VARCHAR(100) NOT NULL,
			description TEXT NOT NULL,
			image VARCHAR(255),
			views INTEGER DEFAULT 0,
			author_id INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)
	`

	_, err = DB.Exec(context.Background(), newsTable)
	if err != nil {
		log.Fatalf("Failed to create news table: %v", err)
	}

	// Ensure image column exists if table was created earlier
	_, _ = DB.Exec(context.Background(), `ALTER TABLE news ADD COLUMN IF NOT EXISTS image VARCHAR(255)`)

	log.Printf("Table news successfully created/checked!")
}
