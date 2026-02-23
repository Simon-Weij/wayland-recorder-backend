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

func InsertVideo(ownerID int, title, videoHash string) (int, error) {
	var id int
	err := database.QueryRow(`
		INSERT INTO videos (owner_id, title, video_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, ownerID, title, videoHash).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
