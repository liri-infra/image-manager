// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	utils "github.com/liri-infra/image-manager/utils"
	"gopkg.in/gcfg.v1"
)

// Settings contains settings from a configuration file.
type Settings struct {
	Server struct {
		Address       string
		SecretKey     string
		UsersDatabase string
	}
	Storage struct {
		Repository string
	}
}

// Application state.
type AppState struct {
	settings *Settings
	logger   *log.Logger
	users    map[string]string
}

func (s *AppState) Settings() *Settings {
	return s.settings
}

func (s *AppState) Logger() *log.Logger {
	return s.logger
}

func (s *AppState) Users() map[string]string {
	return s.users
}

var instance *AppState
var once sync.Once

func GetAppState() *AppState {
	once.Do(func() {
		// Logger
		logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

		// Load settings
		var settingsFileName = "./config.ini"
		if len(os.Args) > 1 {
			settingsFileName = os.Args[1:][0]
		}
		var settings Settings
		logger.Printf("Loading settings from %s\n", settingsFileName)
		err := gcfg.ReadFileInto(&settings, settingsFileName)
		if err != nil {
			panic(err)
		}

		// Expand user
		settings.Server.UsersDatabase = utils.ExpandUser(settings.Server.UsersDatabase)
		settings.Storage.Repository = utils.ExpandUser(settings.Storage.Repository)

		// Open the users database
		usersDatabaseFile, err := os.Open(settings.Server.UsersDatabase)
		if err != nil {
			panic(fmt.Errorf("Unable to open users database %v: %s", settings.Server.UsersDatabase, err.Error()))
		}
		defer usersDatabaseFile.Close()
		usersDatabaseText, _ := ioutil.ReadAll(usersDatabaseFile)
		var users map[string]string
		err = json.Unmarshal(usersDatabaseText, &users)
		if err != nil {
			panic(fmt.Errorf("Failed to unmarshal users database from %v: %s", settings.Server.UsersDatabase, err.Error()))
		}

		// Remove old images
		ticker := time.NewTicker(3600)
		go func() {
			for _ = range ticker.C {
				utils.RemoveOldImages(settings.Storage.Repository)
			}
		}()

		// Create the instance
		instance = &AppState{&settings, logger, users}
	})
	return instance
}
