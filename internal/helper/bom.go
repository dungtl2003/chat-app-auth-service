package helper

import (
	"encoding/json"
	"strings"
	"unicode"
)

func ParseAsJson(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func StripWS(s string) string {
	s = strings.TrimSpace(s)

	var b strings.Builder
	b.Grow(len(s))

	isSpace := false
	for _, ch := range s {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
			isSpace = false
		} else if !isSpace {
			b.WriteRune(' ')
			isSpace = true
		}
	}

	return b.String()
}

func EncodeURLPath(path string) string {
	// Encode the path to make it safe for use in a URL
	return strings.ReplaceAll(path, " ", "%20")
}
