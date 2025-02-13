package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/protected", handleProtectedResource)

	log.Println("OAuth2 server is Running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
