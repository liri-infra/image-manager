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

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Create the directory if needed
	path := filepath.Join(server.GetAppState().Settings().Storage.Repository, params["channel"])
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Read multipart without writing to intermediate files
	reader, err := r.MultipartReader()
	if err != nil {
		server.GetAppState().Logger().Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

		// Destination path
		temp_path := filepath.Join(path, "."+part.FileName()+".part")
		dest_path := filepath.Join(path, part.FileName())

		// Do not allow to overwrite existing files
		if _, err := os.Stat(dest_path); err == nil {
			part.Close()
			server.GetAppState().Logger().Printf("Client tried to overwrite %v\n", part.FileName())
			http.Error(w, fmt.Sprintf("File %v already exist", part.FileName()), http.StatusBadRequest)
			return
		}

		// Write file for the channel specified
		err = server.WriteFile(temp_path, part)
		if err == nil {
			if err = os.Rename(temp_path, dest_path); err != nil {
				part.Close()
				os.Remove(temp_path)
				msg := fmt.Sprintf("Failed to rename %v to %v: %s", filepath.Base(temp_path), filepath.Base(dest_path), err.Error())
				server.GetAppState().Logger().Println(msg)
				http.Error(w, msg, http.StatusInternalServerError)
			}
		} else {
			part.Close()
			os.Remove(temp_path)
			msg := fmt.Sprintf("Failed to write %v: %s", dest_path, err.Error())
			server.GetAppState().Logger().Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}
}
