package utils

import "github.com/google/uuid"

func GenerateUID() string {
	uid := uuid.New().String()

	return uid
}
