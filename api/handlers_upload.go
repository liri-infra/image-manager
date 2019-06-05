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
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	server "github.com/liri-infra/image-manager/server"
)

func Upload(c server.Context, w http.ResponseWriter, r *http.Request) (int, []byte) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	// Create the directory if needed
	path := filepath.Join(c.Settings().Storage.Repository, params["channel"])
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	reader, err := r.MultipartReader()
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
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

		// Write file for the channel specified
		err = server.WriteFile(filepath.Join(path, part.FileName()), part)
		if err != nil {
			return http.StatusInternalServerError, []byte(err.Error())
		}
	}

	return http.StatusOK, nil
}
