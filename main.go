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
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	api "github.com/liri-infra/image-manager/api"
	server "github.com/liri-infra/image-manager/server"
)

func use(handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func main() {
	// Create state
	state := server.GetAppState()

	// Router
	router := mux.NewRouter()
	router.HandleFunc("/jwt/signin", api.SignInHandler).Methods("POST")
	router.HandleFunc("/jwt/refresh", api.RefreshTokenHandler).Methods("POST")
	router.HandleFunc("/api/v1/upload/{channel}", use(api.UploadHandler, api.AuthenticationMiddleware)).Methods("POST")

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(state.Settings().Server.Address, handlers.CompressHandler(loggedRouter))
}
