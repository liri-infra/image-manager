// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"io"
	"os"
)

func WriteFile(filename string, input_file io.Reader) error {
	output_file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer output_file.Close()

	if _, err := io.Copy(output_file, input_file); err != nil {
		return err
	}

	return nil
}
