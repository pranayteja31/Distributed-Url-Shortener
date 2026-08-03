package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func DB_connection(){

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to system environment variables")
	}

	// 2. Fetch the environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	pass := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",host, user, password, dbname, port,pass)

	// 4. Open and verify the connection
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL using environment variables!")

	// queryBytes, err := os.ReadFile("migrations/000001_create_urls.up.sql")
	// if err != nil {
	// 	log.Fatalf("failed to read the sql file: %s", err)
	// }
	// createTableSQL := string(queryBytes)

	// // 3. Execute the SQL statements
	// _, err = DB.Exec(createTableSQL)
	// if err != nil {
	// 	log.Fatalf("Failed to execute SQL script: %v", err)
	// }

	// fmt.Println("SQL file executed successfully!")
}