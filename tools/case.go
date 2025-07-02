package tools

import (
	bin "github.com/gagliardetto/binary"
)

func ToCamelUpper(s string) string {
	return ToCamel(bin.ToRustSnakeCase(s))
}

func toCamelOld(s string) string {
	return ToCamel(s)
}

func ToCamelLower(s string) string {
	return ToLowerCamel(bin.ToRustSnakeCase(s))
}
