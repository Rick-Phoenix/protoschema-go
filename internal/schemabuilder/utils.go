package schemabuilder

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func IsTitleCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstLetter := rune(s[0])

	return unicode.IsUpper(firstLetter)
}

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

func JoinIntSlice(s []int, separator string) string {
	out := ""

	for i, n := range s {
		out += strconv.Itoa(n)
		if i != len(s)-1 {
			out += separator
		}
	}

	return out
}

func IndentLines(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Fprintf(writer, "%s%s\n", indent, line)
		if err != nil {
			return fmt.Errorf("failed to write indented line: %w", err)
		}
	}
	return scanner.Err()
}

func IndentString(s string) string {
	var sb strings.Builder
	err := IndentLines(strings.NewReader(s), &sb)
	if err != nil {
		fmt.Printf("Internal error indenting string: %v\n", err)
		return s // Fallback to unindented string
	}
	return sb.String()
}

func IndentErrors(errs Errors) error {
	if len(errs) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("%s%s\n", indent, IndentString(err.Error())))
	}
	return errors.New(sb.String())
}

func formatProtoValue(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v), nil
	case []byte:
		byteString, err := formatBytesAsProtoLiteral(v)
		if err != nil {
			return "", err
		}
		return byteString, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case *durationpb.Duration:
		return fmt.Sprintf("{ seconds: %d, nanos: %d }", v.GetSeconds(), v.GetNanos()), nil
	case *timestamppb.Timestamp:
		return fmt.Sprintf("{ seconds: %d, nanos: %d }", v.GetSeconds(), v.GetNanos()), nil
	default:
		return "", fmt.Errorf("unsupported type for Protobuf literal formatting: %v", reflect.TypeOf(value))
	}
}

func formatBytesAsProtoLiteral(b []byte) (string, error) {
	var buf bytes.Buffer
	buf.WriteByte('"')
	var err error
	for _, char := range b {
		if char >= 32 && char <= 126 && char != '\\' && char != '"' {
			err = buf.WriteByte(char)
		} else {
			_, err = buf.WriteString(fmt.Sprintf("\\x%02x", char))
		}
	}
	buf.WriteByte('"')

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func formatProtoList[T any](l []T) (string, error) {
	var sb strings.Builder
	sb.WriteString("[")
	for _, v := range l {
		protoVal, err := formatProtoValue(v)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("%s, ", protoVal))
	}
	sb.WriteString("]")

	return sb.String(), nil
}

func formatProtoDict(d map[string]any) (string, error) {
	if len(d) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("{\n")
	for k, v := range d {
		protoV, err := formatProtoValue(v)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, protoV))
	}
	sb.WriteString("}")

	return sb.String(), nil

}

func ValidateDurationString(durationStr string) error {
	_, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("Invalid duration format '%s': %w", durationStr, err)
	}
	return nil
}
