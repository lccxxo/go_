package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, exists := users[username]
	if !exists || user.PassWord != password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := generateAccessToken(user.ID)

	response := map[string]string{"access_token": token.Token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateAccessToken(userID string) AccessToken {
	token := AccessToken{
		Token:     "token_" + userID + "_" + time.Now().Format(time.RFC3339),
		UserId:    userID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tokenMutex.Lock()
	tokens[token.Token] = token
	tokenMutex.Unlock()

	return token
}

func handleProtectedResource(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if accessToken, exists := tokens[token]; exists && accessToken.ExpiresAt.After(time.Now()) {
		w.Write([]byte("Hello, " + accessToken.UserId + "! You have accessed a protected resource."))
	} else {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	}
}
