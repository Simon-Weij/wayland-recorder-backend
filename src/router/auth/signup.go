/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package auth

import (
	"fmt"
	"simon-weij/wayland-recorder-backend/src/database"
	"simon-weij/wayland-recorder-backend/src/dto"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"golang.org/x/crypto/bcrypt"
)

// /auth/signup
func Signup(ctx fiber.Ctx) error {
	var dto dto.User

	if err := ctx.Bind().Body(&dto); err != nil {
		return fiber.ErrBadRequest
	}

	if err := checkForConflicts("email", dto.Email); err != nil {
		return err
	}
	if err := checkForConflicts("username", dto.Username); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	dto.Password = string(hash)

	if err != nil {
		log.Warn("Couldn't hash the password!")
		return fiber.NewError(fiber.ErrInternalServerError.Code, "couldn't hash password")
	}

	database.InsertUserIntoDatabase(dto)

	return nil
}

func checkForConflicts(whatExists string, value string) error {
	userWithEmailExists, err := database.ValueAlreadyExists(whatExists, value)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't check if %s already exists", whatExists))
		return fiber.NewError(fiber.ErrInternalServerError.Code, "couldn't check if %s already exists", whatExists)
	}
	if userWithEmailExists {
		return fiber.NewError(fiber.StatusConflict, fmt.Sprintf("user with %s %s already exists.", whatExists, value))
	}

	return nil
}
