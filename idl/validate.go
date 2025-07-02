package idl

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

func (idlObj *Idl) Validate() *ValidationErrors {
	errs := ValidateIDL(idlObj)
	if errs != nil {
		return errs
	}
	return nil
}

type Strings []string

// Has checks if the string is in the slice.
func (s Strings) Has(str string) bool {
	return slices.Contains(s, str)
}

// AddUnique adds a string to the slice if it is not already present.
func (s *Strings) AddUnique(str string) {
	if !s.Has(str) {
		*s = append(*s, str)
	}
}

// Push adds one or more strings to the slice regardless of whether it is already present.
func (s *Strings) Push(str ...string) {
	*s = append(*s, str...)
}

// Sort sorts the slice of strings.
func (s *Strings) Sort() *Strings {
	if s == nil {
		return nil
	}
	sort.Strings(*s)
	return s
}

// NotIn returns a new Strings containing elements from `a` that are not in `b`.
// It is the opposite of In.
// For example, if a = ["a", "b", "c"] and b = ["b", "c"], then a.NotIn(b) will return ["a"].
// If `b` is nil, it returns a clone of `a`.
// If `a` is nil, it returns an empty Strings.
// If both are nil, it returns an empty Strings.
func (a Strings) NotIn(b Strings) Strings {
	if b == nil || len(b) == 0 {
		return a.Clone()
	}
	if a == nil || len(a) == 0 {
		return Strings{}
	}
	var result Strings
	for _, v := range a {
		if !b.Has(v) {
			result.Push(v)
		}
	}
	return result
}

// In returns a new Strings containing elements that are in both `a` and `b`.
// For example, if a = ["a", "b", "c"] and b = ["b", "c", "d"], then a.In(b) will return ["b", "c"].
// If `b` is nil, it returns an empty Strings.
// If `a` is nil, it returns an empty Strings.
// If both are nil, it returns an empty Strings.
func (a Strings) In(b Strings) Strings {
	if b == nil || len(b) == 0 {
		return Strings{}
	}
	if a == nil || len(a) == 0 {
		return Strings{}
	}
	var result Strings
	for _, v := range a {
		if b.Has(v) {
			result.Push(v)
		}
	}
	return result
}

// Equal checks if two Strings are equal, meaning they contain the same elements regardless of order.
func (s Strings) Equal(other Strings) bool {
	if len(s) != len(other) {
		return false
	}
	for _, v := range s {
		if !other.Has(v) {
			return false
		}
	}
	return true
}

// Clone creates a new Strings with the same elements as the original.
func (s Strings) Clone() Strings {
	result := make(Strings, len(s))
	copy(result, s)
	return result
}

func (s Strings) String() string {
	if len(s) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	result += "]"
	return result
}

func (s Strings) Len() int {
	return len(s)
}

func (s Strings) Duplicates() Strings {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var duplicates Strings
	for _, v := range s {
		if seen[v] {
			duplicates.AddUnique(v)
		} else {
			seen[v] = true
		}
	}
	return duplicates
}

func (s Strings) Unique() Strings {
	if len(s) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var unique Strings
	for _, v := range s {
		if !seen[v] {
			unique.Push(v)
			seen[v] = true
		}
	}
	return unique
}

type ValidationErrors struct {
	NotResolvedTypes          Strings
	InvalidTypeNames          Strings
	DuplicateDefinedTypeNames Strings
	OtherErrors               []error // You can add more error types if needed.
}

func (v *ValidationErrors) IsNil() bool {
	if v == nil {
		return true
	}
	return len(v.NotResolvedTypes) == 0 &&
		len(v.InvalidTypeNames) == 0 &&
		len(v.DuplicateDefinedTypeNames) == 0 &&
		len(v.OtherErrors) == 0
}

func (v *ValidationErrors) Error() string {
	if v.IsNil() {
		return ""
	}
	var result string
	if len(v.NotResolvedTypes) > 0 {
		result += "Not resolved types: " + v.NotResolvedTypes.String() + "\n"
	}
	if len(v.InvalidTypeNames) > 0 {
		result += "Invalid type names: " + v.InvalidTypeNames.String() + "\n"
	}
	if len(v.DuplicateDefinedTypeNames) > 0 {
		result += "Duplicate defined type names: " + v.DuplicateDefinedTypeNames.String() + "\n"
	}
	if len(v.OtherErrors) > 0 {
		result += "Other errors:\n"
		for _, err := range v.OtherErrors {
			result += "- " + err.Error() + "\n"
		}
	}
	return result
}

func (v ValidationErrors) HasErrors() bool {
	return !v.IsNil()
}

func (v *ValidationErrors) AddNotResolvedType(name string) {
	if v.NotResolvedTypes == nil {
		v.NotResolvedTypes = Strings{}
	}
	v.NotResolvedTypes.AddUnique(name)
}

func (v *ValidationErrors) AddInvalidTypeName(name string) {
	if v.InvalidTypeNames == nil {
		v.InvalidTypeNames = Strings{}
	}
	v.InvalidTypeNames.AddUnique(name)
}

func (v *ValidationErrors) AddDuplicateDefinedTypeName(name string) {
	if v.DuplicateDefinedTypeNames == nil {
		v.DuplicateDefinedTypeNames = Strings{}
	}
	v.DuplicateDefinedTypeNames.AddUnique(name)
}

// AddOtherError adds an error to the OtherErrors slice.
func (v *ValidationErrors) AddOtherError(err error) {
	if v.OtherErrors == nil {
		v.OtherErrors = []error{}
	}
	v.OtherErrors = append(v.OtherErrors, err)
}

func ValidateIDL(idl *Idl) *ValidationErrors {
	errs := &ValidationErrors{}

	{
		// check for not resolved types
		wantedTypes := new(Strings)
		definedTypes := new(Strings)
		{
			// each account name is a type we want to resolve
			for _, account := range idl.Accounts {
				if account.Name != "" {
					wantedTypes.AddUnique(account.Name)
				}
			}
			// each event name is a type we want to resolve
			for _, event := range idl.Events {
				if event.Name != "" {
					wantedTypes.AddUnique(event.Name)
				}
			}
		}
		idl.walk_IdlType(func(path pathElements, idlType idltype.IdlType) bool {
			if idlType == nil {
				// This is a type that is not resolved.
				errs.AddNotResolvedType(path.String())
				return true
			}

			// Collect all defined types.
			if definedType, ok := idlType.(*idltype.Defined); ok {
				definedTypes.AddUnique(definedType.Name)
			}

			// Collect all wanted types.
			wantedTypes.Push(getNamedTypeNamesFromIdlType(idlType)...)

			return true
		})
		{
			for _, definedType := range idl.Types {
				if definedType.Name == "" {
					continue // skip unnamed types
				}
				definedTypes.AddUnique(definedType.Name)
			}
		}

		notResolved := wantedTypes.Sort().NotIn(*definedTypes).Unique()
		if len(notResolved) > 0 {
			errs.NotResolvedTypes = notResolved
		}
		duplicates := definedTypes.Duplicates()
		if len(duplicates) > 0 {
			errs.DuplicateDefinedTypeNames = *duplicates.Sort()
		}
		{
			nonValidTypeNames := new(Strings)
			for _, definedType := range *definedTypes {
				if !isValidName(definedType) {
					nonValidTypeNames.AddUnique(definedType)
				}
			}
			if len(*nonValidTypeNames) > 0 {
				errs.InvalidTypeNames = *nonValidTypeNames.Sort()
			}
		}
	}
	{
		// check that all account discriminators are 8 bytes long
		for _, account := range idl.Accounts {
			if len(account.Discriminator) != 8 {
				errs.AddOtherError(
					fmt.Errorf(
						"Account %s has invalid discriminator length: %d, expected 8 bytes",
						account.Name,
						len(account.Discriminator),
					),
				)
			}
		}
		// check that all event discriminators are 8 bytes long
		for _, event := range idl.Events {
			if len(event.Discriminator) != 8 {
				errs.AddOtherError(
					fmt.Errorf(
						"Event %s has invalid discriminator length: %d, expected 8 bytes",
						event.Name,
						len(event.Discriminator),
					),
				)
			}
		}
	}
	{
		// cannot have duplicate instruction names
		instructionNames := new(Strings)
		formattedInstructionNames := new(Strings)
		for index, instruction := range idl.Instructions {
			if instruction.Name == "" {
				errs.AddOtherError(
					fmt.Errorf(
						"Instruction at index %d has no name",
						index,
					),
				)
				continue
			}
			instructionNames.AddUnique(instruction.Name)
			formattedInstructionNames.AddUnique(tools.ToCamelUpper(instruction.Name))
		}
		duplicates := instructionNames.Duplicates()
		if len(duplicates) > 0 {
			errs.AddOtherError(
				fmt.Errorf(
					"Duplicate instruction names found: %s",
					duplicates.String(),
				),
			)
		}
		if len(formattedInstructionNames.Duplicates()) > 0 {
			errs.AddOtherError(
				fmt.Errorf(
					"Duplicate formatted instruction names found: %s",
					formattedInstructionNames.Duplicates().String(),
				),
			)
		}
	}
	if errs.IsNil() {
		return nil
	}

	return errs
}

type pathElements []string

func (p pathElements) String() string {
	return "/" + strings.Join(p, "/")
}

func isValidName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		// TODO: check for valid characters
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_:", r) {
			return false
		}
	}
	return true
}

type IdlTypeCallback func(pathElements, idltype.IdlType) bool

// Walk types from idl.types
func (idlObj *Idl) walk_types(path pathElements, callback IdlTypeCallback) bool {
	if len(idlObj.Types) > 0 {
		for index, typ := range idlObj.Types {
			typPath := append(path, "types", fmt.Sprintf("[%v]", index))
			more := walk_IdlTypeDef(
				typPath,
				typ,
				callback,
			)
			if !more {
				return false
			}
		}
	}
	return true
}

func (idlObj *Idl) walk_IdlType(callback IdlTypeCallback) bool {
	// TODO:
	// - traverse the IDL object tree and look for IDLType
	//   and any other type references.
	// - we're looking for finding all the types that are needed.
	path := pathElements{"idl"}
	{
		for index, ins := range idlObj.Instructions {
			insPath := append(path, "instructions", fmt.Sprintf("[%v]", index))
			more := walk_IdlInstruction(
				insPath,
				ins,
				callback,
			)
			if !more {
				return false
			}
			if ins.Returns.IsSome() {
				more = callback(
					append(insPath,
						"returns",
					),
					ins.Returns.Unwrap(),
				)
				if !more {
					return false
				}
			}
		}
	}

	{
		if !idlObj.walk_types(path, callback) {
			return false
		}
	}

	return true
}

func walk_IdlTypeDef(path pathElements, typ IdlTypeDef, callback IdlTypeCallback) bool {
	switch vv := typ.Ty.(type) {
	case *IdlTypeDefTyEnum:
		more := walk_IdlTypeDefTyEnum(
			append(path,
				"type",
			),
			*vv,
			callback,
		)
		if !more {
			return false
		}
	case *IdlTypeDefTyStruct:
		more := walk_IdlTypeDefTyStruct(
			append(path,
				"type",
			),
			*vv,
			callback,
		)
		if !more {
			return false
		}
	default:
		panic(fmt.Errorf("unknown IDLTypeDef.Type: %T", typ.Ty))
	}
	return true
}

func walk_IdlTypeDefTyStruct(path pathElements, typ IdlTypeDefTyStruct, callback IdlTypeCallback) bool {
	return walk_IdlTypeDefStruct(
		append(path,
			"fields",
		),
		typ,
		callback,
	)
}

func walk_IdlTypeDefTyEnum(path pathElements, typ IdlTypeDefTyEnum, callback IdlTypeCallback) bool {
	for index, field := range typ.Variants {
		fieldPath := append(path, "variants", fmt.Sprintf("[%v]", index))
		more := walk_IdlEnumVariant(
			fieldPath,
			field,
			callback,
		)
		if !more {
			return false
		}
	}
	return true
}

func walk_IdlEnumVariant(path pathElements, typ IdlEnumVariant, callback IdlTypeCallback) bool {
	if typ.Fields.IsNone() {
		return false
	}
	switch vv := typ.Fields.Unwrap().(type) {
	case IdlDefinedFieldsNamed:
		more := walk_IdlEnumFieldsNamed(
			append(path,
				"fields",
			),
			vv,
			callback,
		)
		if !more {
			return false
		}
	case IdlDefinedFieldsTuple:
		more := walk_IdlEnumFieldsTuple(
			append(path,
				"fields",
			),
			vv,
			callback,
		)
		if !more {
			return false
		}
	case nil:
		// TODO: handle nil
	default:
		panic(fmt.Errorf("unknown IDLEnumVariant.Fields: %T", typ.Fields))
	}

	return true
}

func walk_IdlEnumFieldsTuple(path pathElements, fields IdlDefinedFieldsTuple, callback IdlTypeCallback) bool {
	for index, field := range fields {
		fieldPath := append(path, "fields", fmt.Sprintf("[%v]", index))
		more := callback(
			fieldPath,
			field,
		)
		if !more {
			return false
		}
	}
	return true
}

func walk_IdlEnumFieldsNamed(path pathElements, fields IdlDefinedFieldsNamed, callback IdlTypeCallback) bool {
	for index, field := range fields {
		fieldPath := append(path, "fields", fmt.Sprintf("[%v]", index), "type")
		more := callback(
			fieldPath,
			field.Ty,
		)
		if !more {
			return false
		}
	}
	return true
}

func walk_IdlTypeDefStruct(path pathElements, fields IdlTypeDefTyStruct, callback IdlTypeCallback) bool {
	switch fields := fields.Fields.(type) {
	case IdlDefinedFieldsNamed:
		return walk_IdlTypeDefStructNamed(
			path,
			fields,
			callback,
		)
	case IdlDefinedFieldsTuple:
		return walk_IdlTypeDefStructTuple(
			path,
			fields,
			callback,
		)
	case nil:
		// No fields, nothing to walk.
		return true
	default:
		panic(fmt.Errorf("unknown IdlTypeDefTyStruct.Fields: %T", fields))
	}
}

func walk_IdlTypeDefStructNamed(
	path pathElements,
	fields IdlDefinedFieldsNamed,
	callback IdlTypeCallback,
) bool {
	for index, field := range fields {
		fieldPath := append(path, "fields", fmt.Sprintf("[%v]", index), "type")
		more := callback(
			fieldPath,
			field.Ty,
		)
		if !more {
			return false
		}
	}
	return true
}

func walk_IdlTypeDefStructTuple(
	path pathElements,
	fields IdlDefinedFieldsTuple,
	callback IdlTypeCallback,
) bool {
	for index, field := range fields {
		fieldPath := append(path, "fields", fmt.Sprintf("[%v]", index))
		more := callback(
			fieldPath,
			field,
		)
		if !more {
			return false
		}
	}
	return true
}

func walk_IdlInstruction(
	path pathElements,
	ins IdlInstruction,
	callback IdlTypeCallback,
) bool {
	for index, arg := range ins.Args {
		more := callback(
			append(path,
				"args",
				fmt.Sprintf("[%v]", index),
			), arg.Ty)
		if !more {
			return false
		}
	}
	return true
}

func getNamedTypeNamesFromIdlType(idlType idltype.IdlType) []string {
	names := []string{}
	switch v := idlType.(type) {
	case *idltype.Bool:
	case *idltype.U8:
	case *idltype.I8:
	case *idltype.U16:
	case *idltype.I16:
	case *idltype.U32:
	case *idltype.I32:
	case *idltype.F32:
	case *idltype.U64:
	case *idltype.I64:
	case *idltype.F64:
	case *idltype.U128:
	case *idltype.I128:
	case *idltype.U256:
	case *idltype.I256:
	case *idltype.Bytes:
	case *idltype.String:
	case *idltype.Pubkey:
	case *idltype.Defined:
		names = append(names, v.Name)
	case *idltype.Option:
		names = append(names, getNamedTypeNamesFromIdlType(v.Option)...)
	case *idltype.COption:
		names = append(names, getNamedTypeNamesFromIdlType(v.COption)...)
	case *idltype.Vec:
		names = append(names, getNamedTypeNamesFromIdlType(v.Vec)...)
	case *idltype.Array:
		names = append(names, getNamedTypeNamesFromIdlType(v.Type)...)
	case *idltype.Generic:
		// names = append(names, v.Generic)
	case nil:
	default:
		panic(fmt.Sprintf("unknown IDLType: %T", v))
	}
	return names
}
