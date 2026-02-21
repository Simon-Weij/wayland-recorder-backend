package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"simon-weij/wayland-recorder-backend/src/dto"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/joho/godotenv"
)

type DatabaseCredentials struct {
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
	dbName     string
}

var database *sql.DB

func InitialiseDatabase() {
	godotenv.Load()
	credentials := DatabaseCredentials{
		dbUser:     os.Getenv("DB_USERNAME"),
		dbPassword: os.Getenv("DB_PASSWORD"),
		dbHost:     os.Getenv("DB_HOST"),
		dbPort:     os.Getenv("DB_PORT"),
		dbName:     os.Getenv("DB_NAME"),
	}

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		credentials.dbUser,
		credentials.dbPassword,
		credentials.dbHost,
		credentials.dbPort,
		credentials.dbName,
	)

	var err error
	database, err = sql.Open("pgx", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	}

	setupUsersTable()
}

func setupUsersTable() {
	database.Query(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`)
}

func InsertUserIntoDatabase(user dto.User) (int, error) {
	var lastInsertId int
	query := `INSERT INTO users (email, username, password_hash) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	err := database.QueryRow(query, user.Email, user.Username, user.Password).Scan(&lastInsertId)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func ValueAlreadyExists(whatExists string, value string) (bool, error) {
	var valueExists bool

	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE %s = $1
		)
	`, whatExists)
	err := database.QueryRow(query, value).Scan(&valueExists)
	return valueExists, err
}
