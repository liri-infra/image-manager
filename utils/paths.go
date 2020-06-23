// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandUser(path string) string {
	new_path := os.ExpandEnv(path)
	home_dir, _ := os.UserHomeDir()

	if new_path == "~" {
		return home_dir
	} else if strings.HasPrefix(new_path, "~/") {
		return filepath.Join(home_dir, new_path[2:])
	}

	return new_path
}
