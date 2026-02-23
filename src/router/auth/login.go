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
	body, err := parseAndValidate(ctx)
	if err != nil {
		return err
	}

	user, err := authenticate(body)
	if err != nil {
		return err
	}

	token, refreshToken, err := issueTokens(user.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func parseAndValidate(ctx fiber.Ctx) (*dto.LoginRequest, error) {
	var body dto.LoginRequest

	// Validate JSON structure
	if err := ctx.Bind().Body(&body); err != nil {
		return nil, fiber.ErrBadRequest
	}

	// Validate required fields
	if body.Email == "" || body.Password == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Email and password are required")
	}

	return &body, nil
}

func authenticate(body *dto.LoginRequest) (*dto.UserAuth, error) {
	// Fetch user id + password_hash
	user, err := database.GetUserAuthByEmail(body.Email)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't get the user of %s", body.Email))
		return nil, fiber.ErrInternalServerError
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(body.Password),
	); err != nil {
		log.Warn("Couldn't compare hash")
		return nil, fiber.ErrInternalServerError
	}

	return user, nil
}
func issueTokens(userID int) (string, string, error) {
	// Generate JWT
	token, err := GenerateToken(userID)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't generate token for %v", userID))
		return "", "", fiber.ErrInternalServerError
	}

	// Create refresh token
	refresh_token, err := database.CreateRefreshToken(userID, 7*24*time.Hour)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't create refresh token for %v with error %v", userID, err))
		return "", "", fiber.ErrInternalServerError
	}

	return token, refresh_token, nil
}
