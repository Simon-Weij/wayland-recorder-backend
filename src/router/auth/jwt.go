/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package auth

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	jwt "github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func Middleware(ctx fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")

	// Checks if header exists
	if authHeader == "" {
		return fiber.ErrUnauthorized
	}

	// Checks if header matches
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return fiber.ErrUnauthorized
	}

	// Verifies jwt token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return jwtSecret, nil
	})
	if err != nil {
		return fiber.ErrUnauthorized
	}

	// Checks if token is expired
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return fiber.ErrUnauthorized
			}
		}

		ctx.Locals("userID", claims["sub"])
	}

	return ctx.Next()
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
