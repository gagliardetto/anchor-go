package sighash

import (
	bin "github.com/gagliardetto/binary"
)

func ToSnakeForSighash(s string) string {
	return bin.ToRustSnakeCase(s)
}

func ToRustSnakeCase(s string) string {
	return bin.ToRustSnakeCase(s)
}
