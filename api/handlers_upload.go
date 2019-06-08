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

package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	server "github.com/liri-infra/image-manager/server"
)

func Upload(c server.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	params := mux.Vars(r)

	// Create the directory if needed
	path := filepath.Join(c.Settings().Storage.Repository, params["channel"])
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var status_code int = http.StatusOK
	var status_message error = nil

	reader, err := r.MultipartReader()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	for {
		// Read next part, stop at the end of file
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		// Skip empty file names
		if part.FileName() == "" {
			continue
		}

		// Don't continue if we had an error reading the previous part
		if status_code != http.StatusOK {
			continue
		}

		// Destination path
		dest_path := filepath.Join(path, part.FileName())

		// Do not allow to overwrite existing files
		if _, err := os.Stat(dest_path); err == nil {
			c.Logger().Printf("Client tried to overwrite %v\n", part.FileName())

			// Keep reading the next part to avoid interrupting the connection
			// with the client but make sure one all parts are read we exit
			// with an error
			status_code = http.StatusPreconditionFailed
			status_message = fmt.Errorf("File %v already exist", part.FileName())
			return status_code, status_message
			//continue
		}

		// Write file for the channel specified
		err = server.WriteFile(dest_path, part)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return status_code, status_message
}
