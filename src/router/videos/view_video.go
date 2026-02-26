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
	"os"
	"path/filepath"
	"simon-weij/wayland-recorder-backend/src/database"
	"simon-weij/wayland-recorder-backend/src/router/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func ServeVideoById(ctx fiber.Ctx) error {
	var req struct {
		Id int `params:"id"`
	}

	if err := ctx.Bind().URI(&req); err != nil {
		return fiber.ErrBadRequest
	}

	uid, err := auth.GetUID(ctx)
	if err != nil {
		return err
	}

	video, err := database.GetVideoByID(req.Id, uid)
	if video == nil {
		log.Warn("Video was nil!")
		return fiber.ErrInternalServerError
	}
	if err != nil {
		log.Warn(err)
		return fiber.ErrInternalServerError
	}

	videoPath := getVideoFromHash(video.VideoHash, video.Extension)

	return ctx.SendFile(videoPath)
}

func getVideoFromHash(hash string, extension string) string {
	uploadDir := os.Getenv("UPLOAD_DIR")

	filePath := filepath.Join(
		uploadDir,
		string(hash[0]),
		string(hash[1]),
		string(hash[2]),
		string(hash[3]),
		string(hash)+string(extension),
	)
	return filePath
}
