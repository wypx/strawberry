package helper

import (
	"fmt"
	"testing"
)

func TestMytest(t *testing.T) {
	password := encryption{"q5920868"}
	fmt.Println(password.EncryptSHA256())

	t.Skipped()
}
