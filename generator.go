package main

import (
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
			st.Add(genTypeName(opt.Option))
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

func marshalerSignature() Code {
	return Id("MarshalWithEncoder").
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
		)
}

func unmarshalerSignature() Code {
	return Id("UnmarshalWithDecoder").
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
		)
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
							false,
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
							false,
						))
				}

				st.Add(code.Line().Line())
			}
		}

	case IdlTypeDefTyKindEnum:
		code := newStatement()
		enumTypeName := def.Name

		if def.Type.Variants.IsAllUint8() {
			code.Type().Id(enumTypeName).Qual(PkgDfuseBinary, "BorshEnum")
			code.Line().Const().Parens(DoGroup(func(gr *Group) {
				for variantIndex, variant := range def.Type.Variants {

					for docIndex, doc := range variant.Docs {
						if docIndex == 0 {
							gr.Line()
						}
						gr.Comment(doc).Line()
					}

					gr.Id(variant.Name + "_" + enumTypeName).Add(func() Code {
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
						for _, variant := range def.Type.Variants {
							switchBlock.Case(Id(variant.Name + "_" + enumTypeName)).Line().Return(Lit(variant.Name))
						}
						switchBlock.Default().Line().Return(Lit(""))
					})
				})

			// // Declare MarshalWithEncoder
			// code.Line().Line().Func().Params(Id("obj").Id(enumTypeName)).Id("MarshalWithEncoder").
			// 	Params(
			// 		ListFunc(func(params *Group) {
			// 			// Parameters:
			// 			params.Id("encoder").Op("*").Qual(PkgDfuseBinary, "Encoder")
			// 		}),
			// 	).
			// 	Params(
			// 		ListFunc(func(results *Group) {
			// 			// Results:
			// 			results.Err().Error()
			// 		}),
			// 	).
			// 	BlockFunc(func(body *Group) {
			// 		body.Return(Id("encoder").Dot("WriteUint8").Call(Uint8().Call(Id("obj"))))
			// 	})
			// code.Line().Line()

			// // Declare UnmarshalWithDecoder
			// code.Func().Params(Id("obj").Op("*").Id(enumTypeName)).Id("UnmarshalWithDecoder").
			// 	Params(
			// 		ListFunc(func(params *Group) {
			// 			// Parameters:
			// 			params.Id("decoder").Op("*").Qual(PkgDfuseBinary, "Decoder")
			// 		}),
			// 	).
			// 	Params(
			// 		ListFunc(func(results *Group) {
			// 			// Results:
			// 			results.Err().Error()
			// 		}),
			// 	).
			// 	BlockFunc(func(body *Group) {
			// 		id := Id("tpm")
			// 		body.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint8").Call()
			// 		body.Add(ifErrReturnErr())
			// 		body.Id("obj").Op("=").Call(Op("*").Id(enumTypeName)).Call(Op("&").Add(id))
			// 		body.Return(Nil())
			// 	})
			// code.Line().Line()
			st.Add(code.Line())
		} else {
			addTypeNameIsComplexEnum(enumTypeName)
			containerName := formatEnumContainerName(enumTypeName)
			interfaceMethodName := formatInterfaceMethodName(enumTypeName)

			// Declare the interface of the enum type:
			code.Type().Id(enumTypeName).Interface(
				Id(interfaceMethodName).Call(),
				marshalerSignature(),
				unmarshalerSignature(),
			).Line().Line()

			// Declare the enum variants container (non-exported, used internally)
			code.Type().Id(containerName).StructFunc(
				func(structGroup *Group) {
					structGroup.Id("Enum").Qual(PkgDfuseBinary, "BorshEnum").Tag(map[string]string{
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
				if variant.IsUint8() {
					// TODO: make the name {variantTypeName}_{interface_name} ???
					code.Type().Id(variantTypeName).Uint8().Line().Line()
				} else {
					code.Type().Id(variantTypeName).StructFunc(
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
								// TODO: handle tuples
								panic("not handled")
							}
						},
					).Line().Line()
				}

				if variant.IsUint8() {
					// Declare MarshalWithEncoder
					code.Line().Line().Func().Params(Id("obj").Id(variantTypeName)).Id("MarshalWithEncoder").
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
					code.Func().Params(Id("obj").Op("*").Id(variantTypeName)).Id("UnmarshalWithDecoder").
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
								variantTypeName,
								"",
								*variant.Fields.IdlEnumFieldsNamed,
								true,
							))

						// Declare UnmarshalWithDecoder
						code.Line().Line().Add(
							genUnmarshalWithDecoder_struct(
								idl,
								false,
								variantTypeName,
								"",
								*variant.Fields.IdlEnumFieldsNamed,
								bin.TypeID{},
								false,
							))
						code.Line().Line()
					}
				}

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
						enumName := field.Type.GetIdlTypeDefined().Defined
						body.BlockFunc(func(argBody *Group) {
							argBody.List(Id("tmp")).Op(":=").Id(formatEnumContainerName(enumName)).Block()
							argBody.Switch(Id("realvalue").Op(":=").Id("obj").Dot(exportedArgName).Op(".").Parens(Type())).
								BlockFunc(func(switchGroup *Group) {
									// TODO: maybe it's from idl.Accounts ???
									interfaceType := idl.Types.GetByName(enumName)
									for variantIndex, variant := range interfaceType.Type.Variants {
										switchGroup.Case(Op("*").Id(ToCamel(variant.Name))).
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
	isPointer bool,
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
						enumName := field.Type.GetIdlTypeDefined().Defined
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
									for variantIndex, variant := range interfaceType.Type.Variants {

										if variant.IsUint8() {
											// TODO: the actual value is not important;
											//  what's important is the type.
											switchGroup.Case(Lit(variantIndex)).
												BlockFunc(func(caseGroup *Group) {
													caseGroup.Id("obj").Dot(exportedArgName).Op("=").
														Parens(Op("*").Id(ToCamel(variant.Name))).
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
									fieldToDecoderCall(true, "obj", exportedArgName, field.Type.GetIdlTypeOption().Option),
								)
							})
						} else {
							body.Add(
								fieldToDecoderCall(isPointer, "obj", exportedArgName, field.Type),
							)
						}
					}

				}

				body.Return(Nil())
			})
	}
	return code
}

func fieldToDecoderCall(
	isPointer bool,
	receiver string,
	fieldName string,
	idlTypeEnv IdlTypeEnvelope,
) Code {
	body := newStatement()
	switch {
	case idlTypeEnv.IsString():
		{
			body.Add(typeStringToDecoder(isPointer, receiver, fieldName, idlTypeEnv.GetString()))
		}
	case idlTypeEnv.IsIdlTypeOption():
		{
			// TODO: is this case ever reached?
			body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(fieldName))
			body.Line().Add(ifErrReturnErr())
			// opt := idlTypeEnv.GetIdlTypeOption()
			// // TODO: optional = pointer?
			// body.Add(genTypeName(opt.Option))
		}
	case idlTypeEnv.IsIdlTypeVec():
		{
			body.BlockFunc(func(bodyBlock *Group) {
				bodyBlock.List(Id("ln"), Err()).Op(":=").Id("decoder").Dot("ReadLength").Call()
				bodyBlock.Add(ifErrReturnErr())
				vec := idlTypeEnv.GetIdlTypeVec()
				bodyBlock.Id("tmpArr").Op(":=").Make(Index().Add(genTypeName(vec.Vec)), Id("ln"))
				bodyBlock.For(Id("i").Op(":=").Lit(0), Id("i").Op("<").Id("ln"), Id("i").Op("++")).BlockFunc(func(forloop *Group) {
					forloop.Err().Op(":=").Id("tmpArr").Index(Id("i")).Dot("UnmarshalWithDecoder").Call(Id("decoder"))
					forloop.Add(ifErrReturnErr())
				})
				if isPointer {
					bodyBlock.Id(receiver).Dot(fieldName).Op("=").Op("&").Id("tmpArr")
				} else {
					bodyBlock.Id(receiver).Dot(fieldName).Op("=").Id("tmpArr")
				}
			})
			// body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(fieldName))
			// body.Add(ifErrReturnErr())
			// vec := idlTypeEnv.GetIdlTypeVec()
			// body.Index().Add(genTypeName(vec.Vec))
		}
	case idlTypeEnv.IsIdlTypeDefined():
		{
			body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(fieldName))
			// if isPointer  {
			// } else {
			// 	body.Err().Op("=").Id("obj").Dot(fieldName).Dot("UnmarshalWithDecoder").Call(Id("decoder"))
			// }
			body.Line().Add(ifErrReturnErr())
			// body.Add(Id(idlTypeEnv.GetIdlTypeDefined().Defined))
		}
	case idlTypeEnv.IsArray():
		{
			body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(fieldName))
			body.Line().Add(ifErrReturnErr())
			// arr := idlTypeEnv.GetArray()
			// body.Index(Id(Itoa(arr.Num))).Add(genTypeName(arr.Thing))
		}
	default:
		panic(spew.Sdump(idlTypeEnv))
	}

	return body
}

func ifErrReturnErr() Code {
	body := newStatement()
	body.If(Err().Op("!=").Nil()).Block(
		Return(Err()),
	)
	return body
}

func typeStringToDecoder(
	isPointer bool,
	receiver string,
	fieldName string,
	ts IdlTypeAsString,
) Code {
	endianness := Qual("encoding/binary", "LittleEndian")
	stat := newStatement()
	stat.BlockFunc(func(block *Group) {
		switch ts {
		case IdlTypeBool:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadBool").Call()
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadBool").Call()
				block.Add(ifErrReturnErr())
			}
		case IdlTypeU8:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint8").Call()
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadUint8").Call()
				block.Add(ifErrReturnErr())
			}
		case IdlTypeI8:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadInt8").Call()
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadInt8").Call()
				block.Add(ifErrReturnErr())
			}
		case IdlTypeU16:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint16").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadUint16").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeI16:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadInt16").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadInt16").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeU32:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint32").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadUint32").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeI32:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadInt32").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadInt32").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeU64:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint64").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadUint64").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeI64:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadInt64").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadInt64").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeU128:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadUint128").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadUint128").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeI128:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadInt128").Call(endianness)
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadInt128").Call(endianness)
				block.Add(ifErrReturnErr())
			}
		case IdlTypeBytes:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadByte").Call()
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadByte").Call()
				block.Add(ifErrReturnErr())
			}
		case IdlTypeString:
			if isPointer {
				id := Id("tpm" + fieldName)
				block.List(id, Err()).Op(":=").Id("decoder").Dot("ReadString").Call()
				block.Add(ifErrReturnErr())
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.List(Id(receiver).Dot(fieldName), Err()).Op("=").Id("decoder").Dot("ReadString").Call()
				block.Add(ifErrReturnErr())
			}
		case IdlTypePublicKey:
			block.List(Id("buf"), Id("err")).Op(":=").Id("decoder").Dot("ReadNBytes").Call(Lit(32))
			block.If(Err().Op("!=").Nil()).Block(
				Return(Err()),
			)

			if isPointer {
				id := Id("tpm" + fieldName)
				block.Add(id).Op(":=").Qual(PkgSolanaGo, "PublicKeyFromBytes").Call(Id("buf"))
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.Id(receiver).Dot(fieldName).Op("=").Qual(PkgSolanaGo, "PublicKeyFromBytes").Call(Id("buf"))
			}

		// Custom:
		case IdlTypeUnixTimestamp:
			block.List(Id("tmp"), Id("err")).Op(":=").Id("decoder").Dot("ReadInt64").Call(endianness)
			block.If(Err().Op("!=").Nil()).Block(
				Return(Err()),
			)
			if isPointer {
				id := Id("tpm" + fieldName)
				block.Add(id).Op(":=").Qual(PkgSolanaGo, "UnixTimeSeconds").Call(Id("tmp"))
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.Id(receiver).Dot(fieldName).Op("=").Qual(PkgSolanaGo, "UnixTimeSeconds").Call(Id("tmp"))
			}
		case IdlTypeHash:
			// TODO: is it always the same length?
			block.List(Id("buf"), Id("err")).Op(":=").Id("decoder").Dot("ReadNBytes").Call(Lit(32))
			block.If(Err().Op("!=").Nil()).Block(
				Return(Err()),
			)
			if isPointer {
				id := Id("tpm" + fieldName)
				block.Add(id).Op(":=").Qual(PkgSolanaGo, "HashFromBytes").Call(Id("buf"))
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.Id(receiver).Dot(fieldName).Op("=").Qual(PkgSolanaGo, "HashFromBytes").Call(Id("buf"))
			}
		case IdlTypeDuration:
			block.List(Id("tmp"), Id("err")).Op(":=").Id("decoder").Dot("ReadInt64").Call(endianness)
			block.If(Err().Op("!=").Nil()).Block(
				Return(Err()),
			)
			if isPointer {
				id := Id("tpm" + fieldName)
				block.Add(id).Op(":=").Qual(PkgSolanaGo, "DurationSeconds").Call(Id("tmp"))
				block.Id(receiver).Dot(fieldName).Op("=").Op("&").Add(id)
			} else {
				block.Id(receiver).Dot(fieldName).Op("=").Qual(PkgSolanaGo, "DurationSeconds").Call(Id("tmp"))
			}

		default:
			panic(Sf("unknown type string: %s", ts))
		}
	})

	return stat
}
