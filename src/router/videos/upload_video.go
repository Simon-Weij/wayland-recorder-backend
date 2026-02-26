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

package videos

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"simon-weij/wayland-recorder-backend/src/database"
	"simon-weij/wayland-recorder-backend/src/router/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func UploadVideo(ctx fiber.Ctx) error {
	var req struct {
		Title     string `form:"title"`
		IsPrivate *bool  `form:"is_private"`
	}

	if err := ctx.Bind().Form(&req); err != nil {
		return fiber.ErrBadRequest
	}

	if req.Title == "" {
		return fiber.ErrBadRequest
	}

	isPrivate := true
	if req.IsPrivate != nil {
		isPrivate = *req.IsPrivate
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Error("upload FormFile:", err)
		return fiber.ErrInternalServerError
	}

	hashSum, err := calculateHash(fileHeader)
	if err != nil {
		return err
	}

	extension := filepath.Ext(fileHeader.Filename)

	fullLocation := getStoragePath(hashSum, extension)

	if err := saveToDisk(ctx, fileHeader, fullLocation); err != nil {
		return err
	}

	uid, err := auth.GetUID(ctx)
	if err != nil {
		return err
	}

	database.InsertVideo(uid, req.Title, hashSum, extension, isPrivate)

	return ctx.SendString("File uploaded successfully")
}

func calculateHash(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		log.Error("upload open:", err)
		return "", fiber.ErrInternalServerError
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Error("upload hash:", err)
		return "", fiber.ErrInternalServerError
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func getStoragePath(hashSum string, extension string) string {
	uploadLocation := os.Getenv("UPLOAD_DIR")
	firstFolder := filepath.Join(uploadLocation, hashSum[0:1], hashSum[1:2])
	return filepath.Join(firstFolder, hashSum[2:3], hashSum[3:4], hashSum+extension)
}

func saveToDisk(ctx fiber.Ctx, fileHeader *multipart.FileHeader, fullLocation string) error {
	if err := os.MkdirAll(filepath.Dir(fullLocation), 0750); err != nil {
		log.Error("upload mkdir:", err)
		return fiber.ErrInternalServerError
	}
	if err := ctx.SaveFile(fileHeader, fullLocation); err != nil {
		log.Error("upload save:", err)
		return fiber.ErrInternalServerError
	}
	return nil
}
