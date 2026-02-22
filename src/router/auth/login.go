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
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"golang.org/x/crypto/bcrypt"
)

// /auth/login
func Login(ctx fiber.Ctx) error {
	var body dto.LoginRequest

	// Validate JSON structure
	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.ErrBadRequest
	}

	// Validate required fields
	if body.Email == "" || body.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Email and password are required")
	}

	// Fetch user id + password_hash
	user, err := database.GetUserAuthByEmail(body.Email)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't get the user of %s", body.Email))
		return fiber.ErrInternalServerError
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(body.Password),
	); err != nil {
		log.Warn("Couldn't compare hash")
		return fiber.ErrInternalServerError
	}

	// Generate JWT
	token, err := database.GenerateToken(user.ID)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't generate token for %v", user.ID))
		return fiber.ErrInternalServerError
	}

	// Create refresh token
	refresh_token, err := database.CreateRefreshToken(user.ID, 7*24*time.Hour)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't create refresh token for %v with error %v", user.ID, err))
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(fiber.Map{
		"token":         token,
		"refresh_token": refresh_token,
	})
}
