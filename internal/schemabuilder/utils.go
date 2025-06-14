package schemabuilder

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"maps"
	"slices"
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

func MapKeys[T comparable](m map[T]any) []T {
	iter := maps.Keys(m)
	keys := slices.Collect(iter)

	return keys
}

func MapValues[T comparable](m map[any]T) []T {
	iter := maps.Values(m)
	values := slices.Collect(iter)

	return values
}

type Entry[K, V any] struct {
	Key   K
	Value V
}

func MapEntries[K comparable, V any](m map[K]V) []Entry[K, V] {
	iter := maps.All(m)
	var entries []Entry[K, V]

	for k, v := range iter {
		entries = append(entries, Entry[K, V]{Key: k, Value: v})
	}

	return entries

}

func Dedupe[T comparable](s []T) []T {
	seen := make(map[T]struct{})
	var uniqueItems []T

	for _, v := range s {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			uniqueItems = append(uniqueItems, v)
		}
	}

	return uniqueItems
}

func DedupeNonComp[T any](s []T) []T {
	seen := make(map[string]struct{})
	uniqueItems := make([]T, 0, len(s))

	for _, item := range s {
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		err := encoder.Encode(item)
		if err != nil {
			fmt.Printf("Error encoding item %v with gob: %v\n", item, err)
			continue
		}

		fingerprint := buf.String()

		if _, exists := seen[fingerprint]; !exists {
			seen[fingerprint] = struct{}{}
			uniqueItems = append(uniqueItems, item)
		}
	}

	return uniqueItems
}

func FindItem[T any](s []T, filter func(i T) bool) *T {
	idx := slices.IndexFunc(s, filter)
	var item *T

	if idx != -1 {
		item = &s[idx]
	}

	return item
}
