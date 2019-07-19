/****************************************************************************
 * This file is part of Liri.
 *
 * Copyright (C) 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
 *
 * $BEGIN_LICENSE:AGPL3+$
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * $END_LICENSE$
 ***************************************************************************/

package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

var cutoff = 7 * 24 * time.Hour

func RemoveOldImages(repo_path string) {
	now := time.Now()

	err := filepath.Walk(repo_path,
		func(walk_path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				if diff := now.Sub(info.ModTime()); diff > cutoff {
					log.Printf("Deleting %s which is %s old\n", walk_path, diff)
					err := os.Remove(walk_path)
					if err != nil {
						log.Fatal(err.Error())
					}
				}
			}

			return nil
		})
	if err != nil {
		log.Fatal(err.Error())
	}
}
