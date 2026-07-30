package ui

import (
	"os"
	"strings"
)

// EnvOr returns os.Getenv(primary) when set, else os.Getenv(legacy).
func EnvOr(primary, legacy string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(legacy)
}

// EnvTruthy reports whether primary or legacy env is "1" or "true" (case-insensitive).
func EnvTruthy(primary, legacy string) bool {
	v := EnvOr(primary, legacy)
	return v == "1" || strings.EqualFold(v, "true")
}
