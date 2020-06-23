// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	api "github.com/liri-infra/image-manager/api"
	server "github.com/liri-infra/image-manager/server"
)

func main() {
	// Create state
	state := server.GetAppState()

	// Router
	router := mux.NewRouter()

	// JWT Router
	jwtRouter := router.PathPrefix("/jwt").Subrouter()
	jwtRouter.HandleFunc("/signin", api.SignInHandler).Methods("POST")
	jwtRouter.HandleFunc("/refresh", api.RefreshTokenHandler).Methods("POST")

	// API Router
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/upload/{channel}", api.UploadHandler).Methods("POST")
	apiRouter.Use(api.AuthenticationMiddleware)

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	http.ListenAndServe(state.Settings().Server.Address, handlers.CompressHandler(loggedRouter))
}
