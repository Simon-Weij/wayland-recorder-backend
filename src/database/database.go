/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package database

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"simon-weij/wayland-recorder-backend/src/dto"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func GetUserAuthByEmail(email string) (*dto.UserAuth, error) {
	var user dto.UserAuth

	err := database.QueryRow(
		`SELECT id, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.PasswordHash)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateRefreshToken(userID int, duration time.Duration) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	rawToken := hex.EncodeToString(b)

	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(duration)

	query := `
		INSERT INTO refresh_tokens (user_id, hashed_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var tokenID int
	err = database.QueryRow(query, userID, hashedToken, expiresAt).Scan(&tokenID)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return rawToken, nil
}

func RefreshToken(userID int, rawRefreshToken string, tokenDuration time.Duration) (string, error) {
	hash := sha256.Sum256([]byte(rawRefreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	var expiresAt time.Time
	query := `
		SELECT expires_at
		FROM refresh_tokens
		WHERE user_id = $1 AND hashed_token = $2
	`
	err := database.QueryRow(query, userID, hashedToken).Scan(&expiresAt)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	if time.Now().After(expiresAt) {
		return "", errors.New("refresh token expired")
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(tokenDuration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GetUserIDFromRefreshToken(rawToken string) (int, error) {
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	var userID int
	var expiresAt time.Time

	err := database.QueryRow(
		`SELECT user_id, expires_at FROM refresh_tokens WHERE hashed_token = $1`,
		hashedToken,
	).Scan(&userID, &expiresAt)

	if err != nil {
		return 0, err
	}

	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token expired")
	}

	return userID, nil
}

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
