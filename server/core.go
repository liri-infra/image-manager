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

package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

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

		// Create the instance
		instance = &AppState{&settings, logger, users}
	})
	return instance
}
