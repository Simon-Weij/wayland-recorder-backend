/*
 * Wayland recorder is a way to easily make clips and share them.
 * Copyright (C) 2026 Simon-Weij
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package database

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
