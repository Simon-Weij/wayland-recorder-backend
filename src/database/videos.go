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

package database

import (
	"database/sql"
	"time"
)

type Video struct {
	ID        int
	OwnerID   int
	OwnerName string
	Title     string
	VideoHash string
	Extension string
	IsPrivate bool
	CreatedAt time.Time
	IsOwner   bool
}

func InsertVideo(ownerID int, title, videoHash string, extension string, isPrivate bool) (int, error) {
	var id int
	err := database.QueryRow(`
		INSERT INTO videos (owner_id, title, video_hash, extension, is_private)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, ownerID, title, videoHash, extension, isPrivate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetVideoByID(videoID int, currentUserID int) (*Video, error) {
	var video Video

	err := database.QueryRow(`
        SELECT v.id, v.owner_id, u.username, v.title, v.video_hash, v.extension, v.is_private, v.created_at, 
               (v.owner_id = $2) AS is_owner
        FROM videos v
        JOIN users u ON v.owner_id = u.id
        WHERE v.id = $1
    `, videoID, currentUserID).Scan(
		&video.ID,
		&video.OwnerID,
		&video.OwnerName,
		&video.Title,
		&video.VideoHash,
		&video.Extension,
		&video.IsPrivate,
		&video.CreatedAt,
		&video.IsOwner,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &video, nil
}
