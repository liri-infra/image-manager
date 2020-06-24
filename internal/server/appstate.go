// SPDX-FileCopyrightText: 2020 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

// AppState represents the application state.
type AppState struct {
	Config *Config
}

// ContextKey is a type that represent the key of a context.
type ContextKey int

const (
	// KeyAppState is the context key for the app state.
	KeyAppState ContextKey = iota
)
