package utils

import (
	"fmt"
	"unicode"
)

func WithoutSpaces(src string) string {
	r := ""
	for _, c := range src {
		if !unicode.IsSpace(c) {
			r += string(c)
		}
	}

	return r
}

func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
