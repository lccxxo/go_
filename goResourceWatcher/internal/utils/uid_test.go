package utils

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	uid := GenerateUID()
	fmt.Println(uid)
}
