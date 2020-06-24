// SPDX-FileCopyrightText: 2020 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"

	"github.com/liri-infra/image-manager/internal/common"
	"github.com/liri-infra/image-manager/internal/logger"
)

// UploadHandler receives files from the client.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get from context
	ctx := r.Context()
	appState, ok := ctx.Value(KeyAppState).(*AppState)
	if !ok {
		logger.Error("Unable to retrieve app state from context")
		http.Error(w, "no app state found", http.StatusUnprocessableEntity)
		return
	}

	// Get the channel name from the query string
	channelName := chi.URLParam(r, "channel")

	// Channel from configuration
	var imageChannel *ImageChannel
	for _, c := range appState.Config.Channels {
		if c.Name == channelName {
			imageChannel = c
			break
		}
	}
	if imageChannel == nil {
		logger.Errorf("Cannot find \"%s\" channel", channelName)
		http.Error(w, "channel not found", http.StatusNotFound)
		return
	}

	var mr *multipart.Reader
	var part *multipart.Part

	mr, err := r.MultipartReader()
	if err != nil {
		logger.Errorf("Multipart error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save checksums here for later comparison
	checksums := map[string]string{}

	// Read all parts
	for {
		if part, err = mr.NextPart(); err != nil {
			if err == io.EOF {
				// Exit when we read all the parts
				break
			} else {
				logger.Errorf("Error reading part: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if part.FormName() == "file" {
			// Receive file
			fileName := part.FileName()
			logger.Debugf("Receiving \"%s\"...", fileName)

			// Destination path
			destPath := filepath.Join(appState.Config.StorageDir, imageChannel.Path, fileName)
			tempPath := destPath + ".part"

			// Do not allow to overwrite existing files
			if _, err := os.Stat(destPath); os.IsExist(err) {
				err = fmt.Errorf("file \"%s\" already exist", fileName)
				logger.Errorf("Cannot upload: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			// Create the destination file
			file, err := os.Create(tempPath)
			if err != nil {
				logger.Errorf("Unable to create %s: %v", fileName, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Write file and calculate checksum for a verification later
			if _, err = io.Copy(file, part); err != nil {
				logger.Errorf("Failed to copy part to \"%s\": %v", fileName, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			file.Close()
			checksum, err := common.CalculateChecksum(tempPath)
			if err != nil {
				logger.Errorf("Failed to calculate checksum of \"%s\": %v", fileName, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			checksums[fileName] = checksum

			// Rename the temporary file
			if err := os.Rename(tempPath, destPath); err != nil {
				logger.Errorf("Failed to rename \"%s\" to \"%s\": %v", tempPath, destPath, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if part.FormName() == "checksum" {
			// Read checksum calculate by the client
			value := &bytes.Buffer{}
			if _, err := io.Copy(value, part); err != nil {
				logger.Errorf("Failed to read checksum: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			args := strings.Split(value.String(), ":")
			if len(args) != 2 {
				logger.Error("Failed to receive checksum: bad format")
				http.Error(w, "bad checksum format", http.StatusUnprocessableEntity)
				return
			}
			fileName := args[0]
			checksum := args[1]
			if fileName == "" || checksum == "" {
				logger.Error("Failed to receive checksum: empty object name or checksum")
				http.Error(w, "empty object name or checksum", http.StatusUnprocessableEntity)
				return
			}

			// If the checksum doesn't match we remove the file and report the error,
			// so that the next time the file will be uploaded again
			if checksums[fileName] != checksum {
				logger.Errorf("Object \"%s\" has a bad checksum (%s vs %s)", fileName, checksums[fileName], checksum)
				http.Error(w, fmt.Sprintf("bad checksum for %s", fileName), http.StatusUnprocessableEntity)
				return
			}
		} else {
			logger.Errorf("Received unsupported form field %s", part.FormName())
			http.Error(w, fmt.Sprintf("unsupported form field %s", part.FormName()), http.StatusUnprocessableEntity)
			return
		}
	}
}
