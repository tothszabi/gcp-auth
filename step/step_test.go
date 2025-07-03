package step

import (
	"fmt"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "service_account_key_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	path := tmpFile.Name()
	fmt.Println(path)
}
