package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	. "github.com/gagliardetto/utilz"
)

const (
	PkgSolanaGo       = "github.com/gagliardetto/solana-go"
	PkgSolanaGoText   = "github.com/gagliardetto/solana-go/text"
	PkgDfuseBinary    = "github.com/gagliardetto/binary"
	PkgTreeout        = "github.com/gagliardetto/treeout"
	PkgFormat         = "github.com/gagliardetto/solana-go/text/format"
	PkgGoFuzz         = "github.com/gagliardetto/gofuzz"
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
		// TODO: some types have their implementation in github.com/gagliardetto/binary
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
		stat.Index().Byte()
	case IdlTypeString:
		stat.String()
	case IdlTypePublicKey:
		stat.Qual(PkgSolanaGo, "PublicKey")

	// Custom:
	case IdlTypeUnixTimestamp:
		stat.Qual(PkgSolanaGo, "UnixTimeSeconds")
	case IdlTypeHash:
		stat.Qual(PkgSolanaGo, "Hash")
	case IdlTypeDuration:
		stat.Qual(PkgSolanaGo, "DurationSeconds")

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

func genTypeName(idlTypeEnv IdlType) Code {
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
			st.Add(genTypeName(opt.Option))
		}
	case idlTypeEnv.IsIdlTypeVec():
		{
			vec := idlTypeEnv.GetIdlTypeVec()
			st.Index().Add(genTypeName(vec.Vec))
		}
	case idlTypeEnv.IsIdlTypeDefined():
		{
			st.Add(Id(idlTypeEnv.GetIdlTypeDefined().Defined.Name))
		}
	case idlTypeEnv.IsArray():
		{
			arr := idlTypeEnv.GetArray()
			st.Index(Id(Itoa(arr.Num))).Add(genTypeName(arr.Elem))
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

func isComplexEnum(envel IdlType) bool {
	if envel.IsIdlTypeDefined() {
		_, ok := typeRegistryComplexEnum[envel.GetIdlTypeDefined().Defined.Name]
		return ok
	}
	return false
}

func addTypeNameIsComplexEnum(name string) {
	typeRegistryComplexEnum[name] = struct{}{}
}

func registerComplexEnums(idl *IDL, def IdlTypeDef) {
	switch def.Type.Kind {
	case IdlTypeDefTyKindEnum:
		enumTypeName := def.Name
		if !def.Type.Variants.IsSimpleEnum() {
			addTypeNameIsComplexEnum(enumTypeName)
		}
	}
}

func genTypeDef(idl *IDL, withDiscriminator bool, def IdlTypeDef) Code {

	st := newStatement()
	switch def.Type.Kind {
	case IdlTypeDefTyKindStruct:
		code := Empty()
		code.Type().Id(def.Name).StructFunc(func(fieldsGroup *Group) {
			for fieldIndex, field := range *def.Type.Fields {

				for docIndex, doc := range field.Docs {
					if docIndex == 0 && fieldIndex > 0 {
						fieldsGroup.Line()
					}
					fieldsGroup.Comment(doc)
				}
				fieldsGroup.Add(genField(field, field.Type.IsIdlTypeOption())).
					Add(func() Code {
						if field.Type.IsIdlTypeOption() {
							return Tag(map[string]string{
								"bin": "optional",
							})
						}
						return nil
					}())
			}
		})
		st.Add(code.Line())

		{
			// generate encoder and decoder methods (for borsh):
			if GetConfig().Encoding == EncodingBorsh {
				code := Empty()
				exportedAccountName := ToCamel(def.Name)

				toBeHashed := ToCamel(def.Name)

				if withDiscriminator {
					discriminatorName := exportedAccountName + "Discriminator"
					if GetConfig().Debug {
						code.Comment(Sf(`hash("%s:%s")`, bin.SIGHASH_ACCOUNT_NAMESPACE, toBeHashed)).Line()
					}
					sighash := bin.SighashTypeID(bin.SIGHASH_ACCOUNT_NAMESPACE, toBeHashed)
					code.Var().Id(discriminatorName).Op("=").Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
						for _, byteVal := range sighash[:] {
							byteGroup.Lit(int(byteVal))
						}
					}).Op("}")

					// Declare MarshalWithEncoder:
					code.Line().Line().Add(
						genMarshalWithEncoder_struct(
							idl,
							true,
							exportedAccountName,
							discriminatorName,
							*def.Type.Fields,
							true,
						))

					// Declare UnmarshalWithDecoder
					code.Line().Line().Add(
						genUnmarshalWithDecoder_struct(
							idl,
							true,
							exportedAccountName,
							discriminatorName,
							*def.Type.Fields,
							sighash,
						))
				} else {
					// Declare MarshalWithEncoder:
					code.Line().Line().Add(
						genMarshalWithEncoder_struct(
							idl,
							false,
							exportedAccountName,
							"",
							*def.Type.Fields,
							true,
						))

					// Declare UnmarshalWithDecoder
					code.Line().Line().Add(
						genUnmarshalWithDecoder_struct(
							idl,
							false,
							exportedAccountName,
							"",
							*def.Type.Fields,
							bin.TypeID{},
						))
				}

				st.Add(code.Line().Line())
			}
		}

	case IdlTypeDefTyKindEnum:
		code := newStatement()
		enumTypeName := def.Name

		if def.Type.Variants.IsSimpleEnum() {
			code.Type().Id(enumTypeName).Qual(PkgDfuseBinary, "BorshEnum")
			code.Line().Const().Parens(DoGroup(func(gr *Group) {
				for variantIndex, variant := range *def.Type.Variants {

					for docIndex, doc := range variant.Docs {
						if docIndex == 0 {
							gr.Line()
						}
						gr.Comment(doc).Line()
					}

					gr.Id(formatSimpleEnumVariantName(variant.Name, enumTypeName)).Add(func() Code {
						if variantIndex == 0 {
							return Id(enumTypeName).Op("=").Iota()
						}
						return nil
					}()).Line()
				}
				// TODO: check for fields, etc.
			}))

			// Generate stringer for the uint8 enum values:
			code.Line().Line().Func().Params(Id("value").Id(enumTypeName)).Id("String").
				Params().
				Params(String()).
				BlockFunc(func(body *Group) {
					body.Switch(Id("value")).BlockFunc(func(switchBlock *Group) {
						for _, variant := range *def.Type.Variants {
							switchBlock.Case(Id(formatSimpleEnumVariantName(variant.Name, enumTypeName))).Line().Return(Lit(variant.Name))
						}
						switchBlock.Default().Line().Return(Lit(""))
					})

				})
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
					structGroup.Id("Enum").Qual(PkgDfuseBinary, "BorshEnum").Tag(map[string]string{
						"borsh_enum": "true",
					})

					for _, variant := range *def.Type.Variants {
						structGroup.Id(ToCamel(variant.Name)).Id(formatComplexEnumVariantTypeName(enumTypeName, variant.Name))
					}
				},
			).Line().Line()

			for _, variant := range *def.Type.Variants {
				// Name of the variant type if the enum is a complex enum (i.e. enum variants are inline structs):
				variantTypeNameComplex := formatComplexEnumVariantTypeName(enumTypeName, variant.Name)

				// Declare the enum variant types:
				if variant.IsUint8() {
					// TODO: make the name {variantTypeName}_{interface_name} ???
					code.Type().Id(variantTypeNameComplex).Uint8().Line().Line()
				} else {
					code.Type().Id(variantTypeNameComplex).StructFunc(
						func(structGroup *Group) {
							switch {
							case variant.Fields.IdlEnumFieldsNamed != nil:
								for _, variantField := range *variant.Fields.IdlEnumFieldsNamed {
									structGroup.Add(genField(variantField, variantField.Type.IsIdlTypeOption())).
										Add(func() Code {
											if variantField.Type.IsIdlTypeOption() {
												return Tag(map[string]string{
													"bin": "optional",
												})
											}
											return nil
										}())
								}
							default:
								for i, variantTupleItem := range *variant.Fields.IdlEnumFieldsTuple {
									variantField := IdlField{
										Name: fmt.Sprintf("Elem_%d", i),
										Type: variantTupleItem,
									}
									structGroup.Add(genField(variantField, variantField.Type.IsIdlTypeOption())).
										Add(func() Code {
											if variantField.Type.IsIdlTypeOption() {
												return Tag(map[string]string{
													"bin": "optional",
												})
											}
											return nil
										}())
								}
							}
						},
					).Line().Line()
				}

				if variant.IsUint8() {
					// Declare MarshalWithEncoder
					code.Line().Line().Func().Params(Id("obj").Id(variantTypeNameComplex)).Id("MarshalWithEncoder").
						Params(
							ListFunc(func(params *Group) {
								// Parameters:
								params.Id("encoder").Op("*").Qual(PkgDfuseBinary, "Encoder")
							}),
						).
						Params(
							ListFunc(func(results *Group) {
								// Results:
								results.Err().Error()
							}),
						).
						BlockFunc(func(body *Group) {
							body.Return(Nil())
						})
					code.Line().Line()

					// Declare UnmarshalWithDecoder
					code.Func().Params(Id("obj").Op("*").Id(variantTypeNameComplex)).Id("UnmarshalWithDecoder").
						Params(
							ListFunc(func(params *Group) {
								// Parameters:
								params.Id("decoder").Op("*").Qual(PkgDfuseBinary, "Decoder")
							}),
						).
						Params(
							ListFunc(func(results *Group) {
								// Results:
								results.Err().Error()
							}),
						).
						BlockFunc(func(body *Group) {
							body.Return(Nil())
						})
					code.Line().Line()
				} else {
					if variant.Fields != nil && variant.Fields.IdlEnumFieldsNamed != nil {
						// Declare MarshalWithEncoder:
						code.Line().Line().Add(
							genMarshalWithEncoder_struct(
								idl,
								false,
								variantTypeNameComplex,
								"",
								*variant.Fields.IdlEnumFieldsNamed,
								true,
							))

						// Declare UnmarshalWithDecoder
						code.Line().Line().Add(
							genUnmarshalWithDecoder_struct(
								idl,
								false,
								variantTypeNameComplex,
								"",
								*variant.Fields.IdlEnumFieldsNamed,
								bin.TypeID{},
							))
						code.Line().Line()
					}
				}

				// Declare the method to implement the parent enum interface:
				if variant.IsUint8() {
					code.Func().Params(Id("_").Op("*").Id(variantTypeNameComplex)).Id(interfaceMethodName).Params().Block().Line().Line()
				} else {
					code.Func().Params(Id("_").Op("*").Id(variantTypeNameComplex)).Id(interfaceMethodName).Params().Block().Line().Line()
				}
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

func genMarshalWithEncoder_struct(
	idl *IDL,
	withDiscriminator bool,
	receiverTypeName string,
	discriminatorName string,
	fields []IdlField,
	checkNil bool,
) Code {
	code := Empty()
	{
		// Declare MarshalWithEncoder
		code.Func().Params(Id("obj").Id(receiverTypeName)).Id("MarshalWithEncoder").
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("encoder").Op("*").Qual(PkgDfuseBinary, "Encoder")
				}),
			).
			Params(
				ListFunc(func(results *Group) {
					// Results:
					results.Err().Error()
				}),
			).
			BlockFunc(func(body *Group) {
				// Body:
				if withDiscriminator && discriminatorName != "" {
					body.Comment("Write account discriminator:")
					body.Err().Op("=").Id("encoder").Dot("WriteBytes").Call(Id(discriminatorName).Index(Op(":")), False())
					body.If(Err().Op("!=").Nil()).Block(
						Return(Err()),
					)
				}

				for _, field := range fields {
					exportedArgName := ToCamel(field.Name)
					if field.Type.IsIdlTypeOption() {
						body.Commentf("Serialize `%s` param (optional):", exportedArgName)
					} else {
						body.Commentf("Serialize `%s` param:", exportedArgName)
					}

					if isComplexEnum(field.Type) {
						enumTypeName := field.Type.GetIdlTypeDefined().Defined.Name
						body.BlockFunc(func(argBody *Group) {
							argBody.List(Id("tmp")).Op(":=").Id(formatEnumContainerName(enumTypeName)).Block()
							argBody.Switch(Id("realvalue").Op(":=").Id("obj").Dot(exportedArgName).Op(".").Parens(Type())).
								BlockFunc(func(switchGroup *Group) {
									// TODO: maybe it's from idl.Accounts ???
									interfaceType := idl.Types.GetByName(enumTypeName)
									for variantIndex, variant := range *interfaceType.Type.Variants {
										variantTypeNameStruct := formatComplexEnumVariantTypeName(enumTypeName, variant.Name)

										switchGroup.Case(Op("*").Id(variantTypeNameStruct)).
											BlockFunc(func(caseGroup *Group) {
												caseGroup.Id("tmp").Dot("Enum").Op("=").Lit(variantIndex)
												caseGroup.Id("tmp").Dot(ToCamel(variant.Name)).Op("=").Op("*").Id("realvalue")
											})
									}
								})

							argBody.Err().Op(":=").Id("encoder").Dot("Encode").Call(Id("tmp"))

							argBody.If(
								Err().Op("!=").Nil(),
							).Block(
								Return(Err()),
							)

						})
					} else {

						if field.Type.IsIdlTypeOption() {
							if checkNil {
								body.BlockFunc(func(optGroup *Group) {
									// if nil:
									optGroup.If(Id("obj").Dot(ToCamel(field.Name)).Op("==").Nil()).Block(
										Err().Op("=").Id("encoder").Dot("WriteBool").Call(False()),
										If(Err().Op("!=").Nil()).Block(
											Return(Err()),
										),
									).Else().Block(
										Err().Op("=").Id("encoder").Dot("WriteBool").Call(True()),
										If(Err().Op("!=").Nil()).Block(
											Return(Err()),
										),
										Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName)),
										If(Err().Op("!=").Nil()).Block(
											Return(Err()),
										),
									)
								})
							} else {
								body.BlockFunc(func(optGroup *Group) {
									// TODO: make optional fields of accounts a pointer.
									// Write as if not nil:
									optGroup.Err().Op("=").Id("encoder").Dot("WriteBool").Call(True())
									optGroup.If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									)
									optGroup.Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName))
									optGroup.If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									)
								})
							}

						} else {
							body.Err().Op("=").Id("encoder").Dot("Encode").Call(Id("obj").Dot(exportedArgName))
							body.If(Err().Op("!=").Nil()).Block(
								Return(Err()),
							)
						}
					}

				}

				body.Return(Nil())
			})
	}
	return code
}

func genUnmarshalWithDecoder_struct(
	idl *IDL,
	withDiscriminator bool,
	receiverTypeName string,
	discriminatorName string,
	fields []IdlField,
	sighash bin.TypeID,
) Code {
	code := Empty()
	{
		// Declare UnmarshalWithDecoder
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id("UnmarshalWithDecoder").
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("decoder").Op("*").Qual(PkgDfuseBinary, "Decoder")
				}),
			).
			Params(
				ListFunc(func(results *Group) {
					// Results:
					results.Err().Error()
				}),
			).
			BlockFunc(func(body *Group) {
				// Body:
				if withDiscriminator && discriminatorName != "" {
					body.Comment("Read and check account discriminator:")
					body.BlockFunc(func(discReadBody *Group) {
						discReadBody.List(Id("discriminator"), Err()).Op(":=").Id("decoder").Dot("ReadTypeID").Call()
						discReadBody.If(Err().Op("!=").Nil()).Block(
							Return(Err()),
						)
						discReadBody.If(Op("!").Id("discriminator").Dot("Equal").Call(Id(discriminatorName).Index(Op(":")))).Block(
							Return(
								Qual("fmt", "Errorf").Call(
									Line().Lit("wrong discriminator: wanted %s, got %s"),
									Line().Lit(Sf("%v", sighash[:])),
									Line().Qual("fmt", "Sprint").Call(Id("discriminator").Index(Op(":"))),
								),
							),
						)
					})
				}

				for _, field := range fields {
					exportedArgName := ToCamel(field.Name)
					if field.Type.IsIdlTypeOption() {
						body.Commentf("Deserialize `%s` (optional):", exportedArgName)
					} else {
						body.Commentf("Deserialize `%s`:", exportedArgName)
					}

					if isComplexEnum(field.Type) {
						// TODO:
						enumName := field.Type.GetIdlTypeDefined().Defined.Name
						body.BlockFunc(func(argBody *Group) {

							argBody.List(Id("tmp")).Op(":=").New(Id(formatEnumContainerName(enumName)))

							argBody.Err().Op(":=").Id("decoder").Dot("Decode").Call(Id("tmp"))

							argBody.If(
								Err().Op("!=").Nil(),
							).Block(
								Return(Err()),
							)

							argBody.Switch(Id("tmp").Dot("Enum")).
								BlockFunc(func(switchGroup *Group) {
									interfaceType := idl.Types.GetByName(enumName)
									for variantIndex, variant := range *interfaceType.Type.Variants {
										variantTypeNameComplex := formatComplexEnumVariantTypeName(enumName, variant.Name)

										if variant.IsUint8() {
											// TODO: the actual value is not important;
											//  what's important is the type.
											switchGroup.Case(Lit(variantIndex)).
												BlockFunc(func(caseGroup *Group) {
													caseGroup.Id("obj").Dot(exportedArgName).Op("=").
														Parens(Op("*").Id(variantTypeNameComplex)).
														Parens(Op("&").Id("tmp").Dot("Enum"))
												})
										} else {
											switchGroup.Case(Lit(variantIndex)).
												BlockFunc(func(caseGroup *Group) {
													caseGroup.Id("obj").Dot(exportedArgName).Op("=").Op("&").Id("tmp").Dot(ToCamel(variant.Name))
												})
										}
									}
									switchGroup.Default().
										BlockFunc(func(caseGroup *Group) {
											caseGroup.Return(Qual("fmt", "Errorf").Call(Lit("unknown enum index: %v"), Id("tmp").Dot("Enum")))
										})
								})

						})
					} else {

						if field.Type.IsIdlTypeOption() {
							body.BlockFunc(func(optGroup *Group) {
								// if nil:
								optGroup.List(Id("ok"), Err()).Op(":=").Id("decoder").Dot("ReadBool").Call()
								optGroup.If(Err().Op("!=").Nil()).Block(
									Return(Err()),
								)
								optGroup.If(Id("ok")).Block(
									Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName)),
									If(Err().Op("!=").Nil()).Block(
										Return(Err()),
									),
								)
							})
						} else {
							body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName))
							body.If(Err().Op("!=").Nil()).Block(
								Return(Err()),
							)
						}
					}

				}

				body.Return(Nil())
			})
	}
	return code
}

func formatComplexEnumVariantTypeName(enumTypeName string, variantName string) string {
	return ToCamel(Sf("%s_%s", enumTypeName, variantName))
}

func formatSimpleEnumVariantName(variantName string, enumTypeName string) string {
	return ToCamel(Sf("%s_%s", enumTypeName, variantName))
}
