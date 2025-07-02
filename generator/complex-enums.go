package generator

import (
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
)

// typeRegistryComplexEnum contains all types that are a complex enum (and thus implemented as an interface).
var typeRegistryComplexEnum = make(map[string]struct{})

func isComplexEnum(envel idltype.IdlType) bool {
	switch vv := envel.(type) {
	case *idltype.Defined:
		_, ok := typeRegistryComplexEnum[vv.Name]
		return ok
	}
	return false
}

func register_TypeName_as_ComplexEnum(name string) {
	typeRegistryComplexEnum[name] = struct{}{}
}

func registerComplexEnums(def idl.IdlTypeDef) {
	switch vv := def.Ty.(type) {
	case *idl.IdlTypeDefTyEnum:
		enumTypeName := def.Name
		if !vv.IsAllSimple() {
			register_TypeName_as_ComplexEnum(enumTypeName)
		}
	case idl.IdlTypeDefTyEnum:
		enumTypeName := def.Name
		if !vv.IsAllSimple() {
			register_TypeName_as_ComplexEnum(enumTypeName)
		}
	}
}
