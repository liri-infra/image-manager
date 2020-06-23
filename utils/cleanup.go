// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

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
