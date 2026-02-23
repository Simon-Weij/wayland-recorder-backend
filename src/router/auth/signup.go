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
	userDto, err := parseAndValidateSignup(ctx)
	if err != nil {
		return err
	}

	if err := checkUserConflicts(userDto); err != nil {
		return err
	}

	if err := processAndStoreUser(userDto); err != nil {
		return err
	}

	return nil
}

func parseAndValidateSignup(ctx fiber.Ctx) (*dto.User, error) {
	var userDto dto.User

	// Validate JSON structure
	if err := ctx.Bind().Body(&userDto); err != nil {
		return nil, fiber.ErrBadRequest
	}

	return &userDto, nil
}

func checkUserConflicts(userDto *dto.User) error {
	// Check if email and username already exists
	if err := checkForConflicts("email", userDto.Email); err != nil {
		return err
	}
	if err := checkForConflicts("username", userDto.Username); err != nil {
		return err
	}
	return nil
}

func processAndStoreUser(userDto *dto.User) error {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(userDto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("Couldn't hash the password!")
		return fiber.NewError(fiber.ErrInternalServerError.Code, "couldn't hash password")
	}
	userDto.Password = string(hash)

	// Insert user into database
	database.InsertUserIntoDatabase(*userDto)
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
