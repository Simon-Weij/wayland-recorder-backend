/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package database

import (
	"fmt"
	"simon-weij/wayland-recorder-backend/src/dto"
)

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
