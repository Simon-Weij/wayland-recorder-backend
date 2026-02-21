/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"log"

	"simon-weij/wayland-recorder-backend/src/database"
	"simon-weij/wayland-recorder-backend/src/router/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Use(logger.New())

	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)

	database.InitialiseDatabase()

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
