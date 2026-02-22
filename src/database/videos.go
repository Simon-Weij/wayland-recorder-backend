/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
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
