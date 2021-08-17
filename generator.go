package main

import (
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	. "github.com/gagliardetto/utilz"
)

const (
	PkgSolanaGo       = "github.com/gagliardetto/solana-go"
	PkgSolanaGoText   = "github.com/gagliardetto/solana-go/text"
	PkgDfuseBinary    = "github.com/dfuse-io/binary"
	PkgTreeout        = "github.com/gagliardetto/treeout"
	PkgFormat         = "github.com/gagliardetto/solana-go/text/format"
	PkgBorshGo        = "github.com/near/borsh-go"
	PkgGoFuzz         = "github.com/google/gofuzz"
	PkgTestifyRequire = "github.com/stretchr/testify/require"
)

type FileWrapper struct {
	Name string
	File *File
}

func typeStringToType(ts IdlTypeAsString) *Statement {
	stat := newStatement()
	switch ts {
	case IdlTypeBool:
		stat.Bool()
	case IdlTypeU8:
		stat.Uint8()
	case IdlTypeI8:
		stat.Int8()
	case IdlTypeU16:
		// TODO: some types have their implementation in github.com/dfuse-io/binary
		stat.Uint16()
	case IdlTypeI16:
		stat.Int16()
	case IdlTypeU32:
		stat.Uint32()
	case IdlTypeI32:
		stat.Int32()
	case IdlTypeU64:
		stat.Uint64()
	case IdlTypeI64:
		stat.Int64()
	case IdlTypeU128:
		stat.Qual(PkgDfuseBinary, "Uint128")
	case IdlTypeI128:
		stat.Qual(PkgDfuseBinary, "Int128")
	case IdlTypeBytes:
		// TODO:
		stat.Qual(PkgDfuseBinary, "HexBytes")
	case IdlTypeString:
		stat.String()
	case IdlTypePublicKey:
		stat.Qual(PkgSolanaGo, "PublicKey")
	default:
		panic(Sf("unknown type string: %s", ts))
	}

	return stat
}

func genField(field IdlField, pointer bool) Code {
	st := newStatement()
	st.Id(ToCamel(field.Name)).
		Add(func() Code {
			if isComplexEnum(field.Type) {
				return nil
			}
			if pointer {
				return Op("*")
			}
			return nil
		}()).
		Add(genTypeName(field.Type))
	return st
}

func genTypeName(idlTypeEnv IdlTypeEnvelope) Code {
	st := newStatement()
	switch {
	case idlTypeEnv.IsString():
		{
			st.Add(typeStringToType(idlTypeEnv.GetString()))
		}
	case idlTypeEnv.IsIdlTypeOption():
		{
			opt := idlTypeEnv.GetIdlTypeOption()
			// TODO: optional = pointer?
			st.Op("*").Add(genTypeName(opt.Option))
		}
	case idlTypeEnv.IsIdlTypeVec():
		{
			vec := idlTypeEnv.GetIdlTypeVec()
			st.Index().Add(genTypeName(vec.Vec))
		}
	case idlTypeEnv.IsIdlTypeDefined():
		{
			st.Add(Id(idlTypeEnv.GetIdlTypeDefined().Defined))
		}
	case idlTypeEnv.IsArray():
		{
			arr := idlTypeEnv.GetArray()
			st.Index(Id(Itoa(arr.Num))).Add(genTypeName(arr.Thing))
		}
	default:
		panic(spew.Sdump(idlTypeEnv))
	}
	return st
}

func codeToString(code Code) string {
	return Sf("%#v", code)
}

// typeRegistryComplexEnum contains all types that are a complex enum (and thus implemented as an interface).
var typeRegistryComplexEnum = make(map[string]struct{})

func isComplexEnum(envel IdlTypeEnvelope) bool {
	if envel.IsIdlTypeDefined() {
		_, ok := typeRegistryComplexEnum[envel.GetIdlTypeDefined().Defined]
		return ok
	}
	return false
}

func addTypeNameIsComplexEnum(name string) {
	typeRegistryComplexEnum[name] = struct{}{}
}

func genTypeDef(def IdlTypeDef) Code {
	st := newStatement()
	switch def.Type.Kind {
	case IdlTypeDefTyKindStruct:
		code := Empty()
		code.Type().Id(def.Name).StructFunc(func(fieldsGroup *Group) {
			for _, field := range *def.Type.Fields {
				fieldsGroup.Add(genField(field, false))
			}
		})

		st.Add(code.Line())
	case IdlTypeDefTyKindEnum:
		code := newStatement()
		enumTypeName := def.Name

		if def.Type.Variants.IsAllUint8() {
			code.Type().Id(enumTypeName).Qual(PkgBorshGo, "Enum")
			code.Line().Const().Parens(DoGroup(func(gr *Group) {
				for variantIndex, variant := range def.Type.Variants {

					gr.Id(variant.Name).Add(func() Code {
						if variantIndex == 0 {
							return Id(enumTypeName).Op("=").Iota()
						}
						return nil
					}()).Line()
				}
				// TODO: check for fields, etc.
			}))
			st.Add(code.Line())
		} else {
			addTypeNameIsComplexEnum(enumTypeName)
			containerName := formatEnumContainerName(enumTypeName)
			interfaceMethodName := formatInterfaceMethodName(enumTypeName)

			// Declare the interface of the enum type:
			code.Type().Id(enumTypeName).Interface(
				Id(interfaceMethodName).Call(),
			).Line().Line()

			// Declare the enum variants container (non-exported, used internally)
			code.Type().Id(containerName).StructFunc(
				func(structGroup *Group) {
					structGroup.Id("Enum").Qual(PkgBorshGo, "Enum").Tag(map[string]string{
						"borsh_enum": "true",
					})

					for _, variant := range def.Type.Variants {
						structGroup.Id(ToCamel(variant.Name)).Id(ToCamel(variant.Name))
					}
				},
			).Line().Line()

			for _, variant := range def.Type.Variants {
				variantTypeName := ToCamel(variant.Name)

				// Declare the enum variant types:
				code.Type().Id(variantTypeName).StructFunc(
					func(structGroup *Group) {
						if variant.IsUint8() {
							structGroup.Id(variantTypeName).Op("*").Uint8()
						} else {
							switch {
							case variant.Fields.IdlEnumFieldsNamed != nil:
								for _, variantField := range *variant.Fields.IdlEnumFieldsNamed {
									// TODO: pointer, or not?
									structGroup.Add(genField(variantField, true))
								}
							default:
								// TODO: handle tuples
								panic("not handled")
							}
						}
					},
				).Line().Line()

				// Declare the method to implement the parent enum interface:
				code.Func().Params(Id("_").Op("*").Id(variantTypeName)).Id(interfaceMethodName).Params().Block().Line().Line()
			}

			st.Add(code.Line().Line())
		}

		// panic(Sf("not implemented: %s", spew.Sdump(def)))
	default:
		panic(Sf("not implemented: %s", spew.Sdump(def.Type.Kind)))
	}
	return st
}

func formatEnumContainerName(enumTypeName string) string {
	return ToLowerCamel(enumTypeName) + "Container"
}

func formatInterfaceMethodName(enumTypeName string) string {
	return "is" + ToCamel(enumTypeName)
}

func formatBuilderFuncName(insExportedName string) string {
	return "New" + insExportedName + "InstructionBuilder"
}
