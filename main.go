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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	api "github.com/liri-infra/image-manager/api"
	server "github.com/liri-infra/image-manager/server"
	utils "github.com/liri-infra/image-manager/utils"
	"gopkg.in/gcfg.v1"
)

// Context of the application
type ctx struct {
	settings *server.Settings
	logger   *log.Logger
}

func (c ctx) Settings() *server.Settings {
	return c.settings
}

func (c ctx) Logger() *log.Logger {
	return c.logger
}

// Application handler
type appHandler struct {
	*ctx
	handler func(server.Context, http.ResponseWriter, *http.Request) (int, error)
}

func (t appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := t.handler(t.ctx, w, r)
	if code != http.StatusOK {
		http.Error(w, err.Error(), code)
		return
	}
}

// Routes
var routes = []struct {
	method  string
	route   string
	handler func(server.Context, http.ResponseWriter, *http.Request) (int, error)
}{
	{"POST", "/api/v1/upload/{channel}", api.Upload},
}

func main() {
	// Load settings
	var settingsFileName = "./config.ini"
	if len(os.Args) > 1 {
		settingsFileName = os.Args[1:][0]
	}
	var settings server.Settings
	fmt.Printf("Loading settings from %s\n", settingsFileName)
	err := gcfg.ReadFileInto(&settings, settingsFileName)
	if err != nil {
		panic(err)
	}

	// Expand user
	settings.Storage.Repository = utils.ExpandUser(settings.Storage.Repository)

	// Logger
	logger := log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

	// Create context
	appContext := &ctx{&settings, logger}

	// Router
	router := mux.NewRouter()

	for _, detail := range routes {
		router.Handle(detail.route, appHandler{appContext, detail.handler}).Methods(detail.method)
	}

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(settings.Server.Address, handlers.CompressHandler(loggedRouter))
}
