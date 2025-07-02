package idl

import (
	"reflect"

	"github.com/gagliardetto/solana-go"
)

func seed() string {
	return "anchor:idl"
}

func IDLAddress(programID solana.PublicKey) (solana.PublicKey, error) {
	base, _, err := solana.FindProgramAddress([][]byte{}, programID)
	if err != nil {
		return solana.PublicKey{}, err
	}
	return solana.CreateWithSeed(base, seed(), programID)
}

func isNillable(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer:
		return true
	case reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}

func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	if isNillable(reflect.TypeOf(v)) && reflect.ValueOf(v).IsNil() {
		return true
	}
	return false
}
