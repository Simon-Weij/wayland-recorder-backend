package videos

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// /videos/upload
func UploadVideo(ctx fiber.Ctx) error {
	uploadLocation := os.Getenv("UPLOAD_DIR")

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		log.Error("upload FormFile:", err)
		return fiber.ErrInternalServerError
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		log.Error("upload open:", err)
		return fiber.ErrInternalServerError
	}
	defer file.Close()

	// Hash file
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Error("upload hash:", err)
		return fiber.ErrInternalServerError
	}
	hashSum := hex.EncodeToString(hasher.Sum(nil))

	// Calculate location
	extension := filepath.Ext(fileHeader.Filename)
	firstFolder := filepath.Join(uploadLocation, hashSum[0:1], hashSum[1:2])
	fullLocation := filepath.Join(firstFolder, hashSum[2:3], hashSum[3:4], hashSum+extension)

	// Save file to location
	if err := os.MkdirAll(filepath.Dir(fullLocation), 0750); err != nil {
		log.Error("upload mkdir:", err)
		return fiber.ErrInternalServerError
	}
	if err := ctx.SaveFile(fileHeader, fullLocation); err != nil {
		log.Error("upload save:", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendString("File uploaded successfully: " + fullLocation)
}
