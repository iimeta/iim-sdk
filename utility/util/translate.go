package util

import (
	"unicode"
)

func HasChinese(text string) bool {

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}

	return false
}
