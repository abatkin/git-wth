package util

import (
	"fmt"
)

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func Contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func Pluralize(n int, word string) string {
	suffix := ""
	if n != 1 {
		suffix = "s"
	}
	return fmt.Sprintf("%d %s%s", n, word, suffix)
}
