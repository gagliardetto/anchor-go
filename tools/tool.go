package tools

import (
	"encoding/json"
	"fmt"
	"go/token"
	"runtime"
	"strings"

	"github.com/tidwall/gjson"
)

func RequireFields(
	dst json.RawMessage,
	fields ...string,
) error {
	for _, field := range fields {
		if !gjson.GetBytes(dst, field).Exists() {
			return fmt.Errorf("missing field %s", field)
		}
	}
	return nil
}

func RequireOneOfFields(
	dst json.RawMessage,
	fields ...string, // at least one of these fields must exist
) error {
	found := false
	for _, field := range fields {
		if gjson.GetBytes(dst, field).Exists() {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("not found any of the required fields %s", strings.Join(fields, ", "))
	}
	return nil
}

func TryUnmarshal[T any](data []byte) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}

func OneOf[T any](data []byte, unmarshalers ...func([]byte) (T, error)) (T, error) {
	for _, unmarshal := range unmarshalers {
		child, err := unmarshal(data)
		if err == nil {
			return child, nil
		}
	}
	return *new(T), fmt.Errorf("no matching child type found for %s from %s", string(data), getCallaPath())
}

func getCallaPath() string {
	paths := _callPath(10)
	if len(paths) == 0 {
		return ""
	}
	return strings.Join(paths, "\n")
}

func _callPath(n int) []string {
	paths := make([]string, 0, n)
	for i := 1; i < n; i++ {
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		file, line := fn.FileLine(pc)
		if strings.HasPrefix(file, "/usr/local/go") {
			break
		}
		paths = append(paths, fmt.Sprintf("%s:%d", file, line))
	}
	return paths
}

func Into[T any](
	dst *T,
	data []byte,
	unmarshalers ...func([]byte) (T, error),
) error {
	_dst, err := OneOf(data, unmarshalers...)
	if err != nil {
		return err
	}
	*dst = _dst
	return nil
}

func IntoArray[T any](
	dst *[]T,
	data []byte,
	unmarshalers ...func([]byte) (T, error),
) error {
	_dst, err := OneOfArray(data, unmarshalers...)
	if err != nil {
		return err
	}
	*dst = _dst
	return nil
}

// same as one of but returns an array of _Child
func OneOfArray[T any](data []byte, unmarshalers ...func([]byte) (T, error)) ([]T, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var children []T
	for _, childData := range raw {
		child, err := OneOf(childData, unmarshalers...)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

func IsReservedKeyword(s string) bool {
	return token.Lookup(s).IsKeyword()
}

func IsValidIdent(s string) bool {
	return token.IsIdentifier(s) && s != "_" // “_” is the blank identifier – forbidden as a package name
}
