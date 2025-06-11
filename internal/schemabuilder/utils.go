package schemabuilder

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"unicode"
)

func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	var chunks []string
	var currentChunk []rune
	for i, letter := range s {
		if i == 0 || !unicode.IsUpper(letter) {
			currentChunk = append(currentChunk, unicode.ToLower(letter))
		} else {
			prevIsLower := unicode.IsLower(rune(s[i-1]))
			nextIsLower := true
			if i != len(s)-1 {
				nextIsLower = unicode.IsLower(rune(s[i+1]))
			}
			if prevIsLower || nextIsLower {
				chunks = append(chunks, string(currentChunk))
				currentChunk = currentChunk[:0]
			}
			currentChunk = append(currentChunk, unicode.ToLower(letter))
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, string(currentChunk))
	}

	return strings.Join(chunks, "_")
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}

	letters := []rune(s)

	letters[0] = unicode.ToUpper(letters[0])

	return string(letters)
}

func RandomString(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	_, err := rand.Read(b) // Read random bytes into the slice
	if err != nil {
		return "", err
	}
	// Encode the random bytes into a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b), nil
}
