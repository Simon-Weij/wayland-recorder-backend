package videos

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"simon-weij/wayland-recorder-backend/src/database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func UploadVideo(ctx fiber.Ctx) error {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Error("upload FormFile:", err)
		return fiber.ErrInternalServerError
	}

	title := ctx.FormValue("title")
	if title == "" {
		return fiber.ErrBadRequest
	}

	hashSum, err := calculateHash(fileHeader)
	if err != nil {
		return err
	}

	fullLocation := getStoragePath(hashSum, filepath.Ext(fileHeader.Filename))

	if err := saveToDisk(ctx, fileHeader, fullLocation); err != nil {
		return err
	}

	uid, err := getUID(ctx)
	if err != nil {
		return err
	}

	database.InsertVideo(uid, title, hashSum)

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

func getUID(ctx fiber.Ctx) (int, error) {
	userID := ctx.Locals("userID")
	uid, ok := userID.(int)
	if !ok {
		log.Error(fmt.Sprintf("Couldn't get user id for %v", userID))
		return 0, fiber.ErrInternalServerError
	}
	return uid, nil
}
