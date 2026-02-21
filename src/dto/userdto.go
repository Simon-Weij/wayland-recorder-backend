/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package dto

type User struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Username string `db:"username"`
	Password string `db:"password_hash"`
}
