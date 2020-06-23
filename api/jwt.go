// SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	server "github.com/liri-infra/image-manager/server"
)

// Struct to read username and password from the request body
type Credentials struct {
	Password string `json:"password"`
	UserName string `json:"username"`
}

// Struct that will be encoded to a JWT
type Claims struct {
	UserName string `json:"username"`
	jwt.StandardClaims
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return an error if the user doesn't exist or the password is wrong
	expectedPassword, ok := server.GetAppState().Users()[creds.UserName]
	if !ok || expectedPassword != creds.Password {
		server.GetAppState().Logger().Printf("User %v doesn't exist or the password is wrong\n", creds.UserName)
		http.Error(w, "Missing user or wrong password", http.StatusUnauthorized)
		return
	}

	// Set expiration time to 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)

	// Create the JWT claims, which includes the username and expiration time
	claims := &Claims{
		UserName: creds.UserName,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString([]byte(server.GetAppState().Settings().Server.SecretKey))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the token to the client
	RespondWithJson(w, tokenString)
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
	cookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tknStr := cookie.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(server.GetAppState().Settings().Server.SecretKey), nil
	})
	if !tkn.Valid {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// (END) The code up-till this point is the same as the first part of the `Welcome` route

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		http.Error(w, "Not enough time has passed", http.StatusBadRequest)
		return
	}

	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(server.GetAppState().Settings().Server.SecretKey))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the token to the client
	RespondWithJson(w, tokenString)
}

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization header and remove what we don't need
		tokenString := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
		if len(tokenString) == 0 {
			server.GetAppState().Logger().Println("Missing authorization header")
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Verify the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(server.GetAppState().Settings().Server.SecretKey), nil
		})
		if token == nil || !token.Valid {
			server.GetAppState().Logger().Println(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				server.GetAppState().Logger().Println(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			server.GetAppState().Logger().Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
