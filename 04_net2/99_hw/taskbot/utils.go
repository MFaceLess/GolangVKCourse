package main

import (
	"strings"
)

func startsWith(src string, substr string) bool {
	return strings.Index(src, substr) == 0
}
