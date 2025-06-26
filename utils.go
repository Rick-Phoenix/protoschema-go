package schemabuilder

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func formatProtoValue[T any](value T) (string, error) {
	switch v := any(value).(type) {
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
	case CelOption:
		return fmt.Sprintf("{\nid: %q \nmessage: %q\nexpression: %q\n}",
			v.Id, v.Message, v.Expression), nil
	default:
		// If it's not one of the direct cases, use reflect to determine its kind.
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		kind := val.Kind()

		if kind == reflect.Slice || kind == reflect.Array { // Handle both slices and arrays identically
			var formattedElements []string
			for i := range val.Len() {
				elem := val.Index(i).Interface()       // Get the element's underlying value as an interface{}
				elemStr, err := formatProtoValue(elem) // Recursively format each element
				if err != nil {
					return "", err
				}
				formattedElements = append(formattedElements, elemStr)
			}
			return fmt.Sprintf("[%s]", strings.Join(formattedElements, ", ")), nil
		} else if kind == reflect.Map {
			var formattedEntries []string
			keys := val.MapKeys()

			sort.Slice(keys, func(i, j int) bool {
				return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
			})

			for _, keyReflectValue := range keys {
				valueReflectValue := val.MapIndex(keyReflectValue)

				keyStr, err := formatProtoValue(keyReflectValue.Interface())
				if err != nil {
					return "", err
				}
				cleanedKey := strings.ReplaceAll(keyStr, "\"", "")
				valStr, err := formatProtoValue(valueReflectValue.Interface())
				if err != nil {
					return "", err
				}
				formattedEntries = append(formattedEntries, fmt.Sprintf("%s: %s", cleanedKey, valStr))
			}
			return fmt.Sprintf("{%s}", strings.Join(formattedEntries, ", ")), nil
		} else if kind == reflect.Struct {
			fields := make(map[string]any)
			typ := val.Type()
			for i := range val.NumField() {
				field := typ.Field(i)
				if field.IsExported() {
					fieldVal := val.Field(i)
					fields[strings.ToLower(field.Name)] = fieldVal.Interface()
				}
			}

			valStr, err := formatProtoValue(fields)
			if err != nil {
				return "", err
			}

			return valStr, nil
		}

		return "", fmt.Errorf("unsupported type for proto value formatting: %T (kind: %s)", value, kind)
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

func addServiceSuffix(name string) string {
	if !strings.HasSuffix(name, "Service") {
		return name + "Service"
	}

	return name
}

func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstLetter := rune(s[0])

	return unicode.IsUpper(firstLetter)
}

func toSnakeCase(s string) string {
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

func capitalize(s string) string {
	if s == "" {
		return s
	}

	letters := []rune(s)

	letters[0] = unicode.ToUpper(letters[0])

	return string(letters)
}

func randomString(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	_, err := rand.Read(b) // Read random bytes into the slice
	if err != nil {
		return "", err
	}
	// Encode the random bytes into a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b), nil
}

func getMapKeys[T comparable](m map[T]any) []T {
	iter := maps.Keys(m)
	keys := slices.Collect(iter)

	return keys
}

func mapValues[T comparable](m map[any]T) []T {
	iter := maps.Values(m)
	values := slices.Collect(iter)

	return values
}

type Entry[K, V any] struct {
	Key   K
	Value V
}

func mapEntries[K comparable, V any](m map[K]V) []Entry[K, V] {
	iter := maps.All(m)
	var entries []Entry[K, V]

	for k, v := range iter {
		entries = append(entries, Entry[K, V]{Key: k, Value: v})
	}

	return entries
}

func dedupe[T comparable](s []T) []T {
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

func dedupeNonComp[T any](s []T) []T {
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

func findItem[T any](s []T, filter func(i T) bool) *T {
	idx := slices.IndexFunc(s, filter)
	var item *T

	if idx != -1 {
		item = &s[idx]
	}

	return item
}

func joinIntSlice(s []int, separator string) string {
	out := ""

	for i, n := range s {
		out += strconv.Itoa(n)
		if i != len(s)-1 {
			out += separator
		}
	}

	return out
}

func joinInt32Slice(s []int32, separator string) string {
	out := ""

	for i, n := range s {
		out += fmt.Sprintf("%d", n)
		if i != len(s)-1 {
			out += separator
		}
	}

	return out
}

func joinUintSlice(s []uint, separator string) string {
	out := ""

	for i, n := range s {
		out += fmt.Sprintf("%d", n)
		if i != len(s)-1 {
			out += separator
		}
	}

	return out
}

func indentString(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	reader := strings.NewReader(text)
	var writer *strings.Builder
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Fprintf(writer, "%s%s\n", indent, line)
		if err != nil {
			return "", fmt.Errorf("Failed to write indented line: %w", err)
		}
	}
	return writer.String(), scanner.Err()
}

func indentErrors(description string, errs error) error {
	if errs == nil {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(description)
	sb.WriteString(":\n")

	indentedErrs, err := indentString(errs.Error())
	if err != nil {
		fmt.Printf("Internal error indenting string: %v\n", err)
		return errs
	}

	sb.WriteString(indentedErrs)

	return errors.New(sb.String())
}

func validateDurationString(durationStr string) error {
	_, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("Invalid duration format '%s': %w", durationStr, err)
	}
	return nil
}

func sliceIntersects[T comparable](s1 []T, s2 []T) bool {
	for _, v := range s1 {
		if slices.Contains(s2, v) {
			return true
		}
	}

	return false
}

func filterAndDedupe[T comparable](target []T, filter func(T) bool) []T {
	seen := make(map[T]struct{})
	out := []T{}

	for _, i := range target {
		if _, alreadySeen := seen[i]; alreadySeen {
			continue
		}

		seen[i] = present

		if filter(i) {
			out = append(out, i)
		}
	}

	return out
}

func mapsMultiCopy[M ~map[K]V, K comparable, V any](dst M, sources ...M) M {
	out := make(M)
	for _, m := range sources {
		maps.Copy(out, m)
	}

	return out
}

func sortString(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}

	return 0
}
