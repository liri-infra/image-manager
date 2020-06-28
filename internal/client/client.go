// SPDX-FileCopyrightText: 2020 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package client

import "github.com/liri-infra/image-manager/internal/logger"

// StartClient starts the client.
func StartClient(url, token string, channel string, paths []string) error {
	// Client
	client, err := NewClient(url, token)
	if err != nil {
		return err
	}

	// Upload
	for _, path := range paths {
		if err := client.UploadSingle(channel, path); err != nil {
			return err
		}
	}

	logger.Info("Done!")

	return nil
}
