package main

import (
	"sync"
	"time"
)

type User struct {
	ID       string
	UserName string
	PassWord string
}

type AccessToken struct {
	Token     string
	UserId    string
	ExpiresAt time.Time
}

var (
	users      = make(map[string]User)
	tokens     = make(map[string]AccessToken)
	tokenMutex sync.Mutex
)

func init() {
	users["testUser"] = User{
		ID:       "1",
		UserName: "testUser",
		PassWord: "password",
	}
}
