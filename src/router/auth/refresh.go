/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package auth

import (
	"fmt"
	"simon-weij/wayland-recorder-backend/src/database"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// /auth/refresh
func RefreshToken(ctx fiber.Ctx) error {
	// Define struct
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Check if the expected format is matched
	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.ErrBadRequest
	}

	// Checks if a refresh token was provided
	if body.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	// Gets user id
	userID, err := database.GetUserIDFromRefreshToken(body.RefreshToken)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	// Generates token
	accessToken, err := GenerateToken(userID)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't generate token for %v", userID))
		return fiber.ErrInternalServerError
	}

	// Generates new refresh token
	newRefreshToken, err := database.CreateRefreshToken(userID, 7*24*time.Hour)
	if err != nil {
		log.Warn(fmt.Sprintf("Couldn't generate refreshtoken for %v", userID))
		return fiber.ErrInternalServerError
	}

	// Returns tokens
	return ctx.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}
