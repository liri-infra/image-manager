// SPDX-FileCopyrightText: 2020 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"os"
	"path/filepath"
	"time"

	"github.com/liri-infra/image-manager/internal/logger"
)

var interval = 7 * 24 * time.Hour

// RemoveOldImages removes images stored inside archivePath for the specified imageChannels.
func RemoveOldImages(archivePath string, imageChannels []*ImageChannel) {
	for _, imageChannel := range imageChannels {
		// Skip those channels that we don't want to clean up
		if !imageChannel.Cleanup {
			continue
		}

		channelPath := filepath.Join(archivePath, imageChannel.Path)
		now := time.Now()

		err := filepath.Walk(channelPath,
			func(walkPath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					if diff := now.Sub(info.ModTime()); diff > interval {
						logger.Infof("Deleting %s which is %s old", walkPath, diff)
						err := os.Remove(walkPath)
						if err != nil {
							return err
						}
					}
				}

				return nil
			})

		if err != nil {
			logger.Errorf("Archive cleanup for channel \"%s\" has failed: %v", imageChannel.Name, err)
		}
	}
}
