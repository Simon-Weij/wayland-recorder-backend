/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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
	setupRefreshTokenTable()
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

func setupRefreshTokenTable() {
	database.Query(`
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		hashed_token VARCHAR(255) NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);
	`)
}
