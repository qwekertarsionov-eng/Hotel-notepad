package config

import (
	"os"
	"strings"
	"testing"
)

func TestTST_INFRA_002(t *testing.T) {
	data, err := os.ReadFile("../../.env.example")
	if err != nil {
		t.Skip(".env.example not found")
	}
	text := string(data)
	for _, key := range []string{"DATABASE_URL", "JWT_SECRET", "HTTP_ADDR"} {
		if !strings.Contains(text, key) {
			t.Errorf(".env.example must contain %s", key)
		}
	}
}
