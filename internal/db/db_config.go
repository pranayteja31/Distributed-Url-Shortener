package db

import (
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var DB *sqlx.DB

func DB_connection() *sqlx.DB {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to system environment variables")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	pass := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, pass)

	dbConn, err := sqlx.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	if err = dbConn.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	DB = dbConn
	fmt.Println("Successfully connected to PostgreSQL using environment variables!")
	return dbConn
}