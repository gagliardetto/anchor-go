package main

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	. "github.com/gagliardetto/utilz"
)

func main() {

	filenames := []string{
		// "idl_files/zero_copy.json",
		// "idl_files/typescript.json",
		// "idl_files/sysvars.json",
		// "idl_files/swap.json",
		"idl_files/swap_light.json",
		// "idl_files/pyth.json",
		// "idl_files/multisig.json",
		// "idl_files/misc.json",
		// "idl_files/lockup.json",
		// "idl_files/ido_pool.json",
		// "idl_files/events.json",
		// "idl_files/escrow.json",
		// "idl_files/errors.json",
		// "idl_files/composite.json",
		// "idl_files/chat.json",
		// "idl_files/cashiers_check.json",
		// "idl_files/counter_auth.json",
		// "idl_files/counter.json",
	}
	for _, idlFilepath := range filenames {
		Ln(LimeBG(idlFilepath))
		idlFile, err := os.Open(idlFilepath)
		if err != nil {
			panic(err)
		}

		dec := json.NewDecoder(idlFile)

		var idl IDL

		err = dec.Decode(&idl)
		if err != nil {
			panic(err)
		}

		spew.Dump(idl)

		err = GenerateClient(idl)
		if err != nil {
			panic(err)
		}
	}
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
		stat.Qual("github.com/dfuse-io/binary", "Uint128")
	case IdlTypeI128:
		stat.Qual("github.com/dfuse-io/binary", "Int128")
	case IdlTypeBytes:
		// TODO:
		stat.Qual("github.com/dfuse-io/binary", "HexBytes")
	case IdlTypeString:
		stat.String()
	case IdlTypePublicKey:
		stat.Qual("github.com/gagliardetto/solana-go", "PublicKey")
	default:
		panic(Sf("unknown type string: %s", ts))
	}

	return stat
}

func idlTypeToType(envel IdlTypeEnvelope) *Statement {
	switch {
	case envel.IsString():
		return typeStringToType(envel.GetString())
	case envel.IsIdlTypeDefined():
		return Id(envel.GetIdlTypeDefined().Defined)
	default:
		panic(spew.Sdump(envel))
	}
}

func GenerateClient(idl IDL) error {
	// TODO:
	// - validate IDL (???)
	// - create new go file
	// - add instructions, etc.

	file := NewGoFile(idl.Name, true)
	for _, programDoc := range idl.Docs {
		file.HeaderComment(programDoc)
	}

	// Instruction ID enum:
	{
		code := Empty()
		code.Const().Parens(
			DoGroup(func(gr *Group) {
				for instructionIndex, instruction := range idl.Instructions {
					insExportedName := ToCamel(instruction.Name)

					ins := Empty().Line()
					for _, doc := range instruction.Docs {
						ins.Comment(doc).Line()
					}
					ins.Id("Instruction_" + insExportedName)
					if instructionIndex == 0 {
						ins.Uint32().Op("=").Iota().Line()
					}
					gr.Add(ins.Line().Line())
				}
			}),
		)
		file.Add(code.Line())
	}

	{
		{ // Base Instruction struct:
			code := Empty()
			code.Type().Id("Instruction").Struct(
				Qual("github.com/dfuse-io/binary", "BaseVariant"),
			)
			file.Add(code.Line())
		}
		{
			// PROGRAM_ID variable:
			code := Empty()
			// TODO: add this to IDL???
			programID := "TODO"
			code.Var().Id("PROGRAM_ID").Op("=").Qual("github.com/gagliardetto/solana-go", "MustPublicKeyFromBase58").Call(Lit(programID))
			file.Add(code.Line())
		}
		{
			// register decoder:
			code := Empty()
			code.Func().Id("init").Call().Block(
				Qual("github.com/gagliardetto/solana-go", "RegisterInstructionDecoder").Call(Id("PROGRAM_ID"), Id("registryDecodeInstruction")),
			)
			file.Add(code.Line())
		}
		{
			// variant definitions for the decoder:
			code := Empty()
			code.Var().Id("InstructionImplDef").Op("=").Qual("github.com/dfuse-io/binary", "NewVariantDefinition").
				Parens(DoGroup(func(call *Group) {
					call.Line()
					// TODO: make this configurable?
					call.Qual("github.com/dfuse-io/binary", "Uint32TypeIDEncoding").Op(",").Line()

					call.Index().Qual("github.com/dfuse-io/binary", "VariantType").
						BlockFunc(func(variantBlock *Group) {
							for _, instruction := range idl.Instructions {
								insName := ToSnake(instruction.Name)
								insExportedName := ToCamel(instruction.Name)
								variantBlock.Block(
									List(Lit(insName), Parens(Op("*").Id(insExportedName)).Parens(Nil())).Op(","),
								).Op(",")
							}
						}).Op(",").Line()
				}))

			file.Add(code.Line())
		}
		{
			// method to return programID:
			code := Empty()
			code.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("ProgramID").Params().
				Parens(Qual("github.com/gagliardetto/solana-go", "PublicKey")).
				BlockFunc(func(gr *Group) {
					gr.Return(
						Id("PROGRAM_ID"),
					)
				})
			file.Add(code.Line())
		}
		{
			// method to return accounts:
			code := Empty()
			code.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("Accounts").Params().
				Parens(Id("out").Index().Op("*").Qual("github.com/gagliardetto/solana-go", "AccountMeta")).
				BlockFunc(func(body *Group) {
					body.Return(
						Id("inst").Dot("Impl").Op(".").Parens(Qual("github.com/gagliardetto/solana-go", "AccountsGettable")).Dot("GetAccounts").Call(),
					)
				})
			file.Add(code.Line())
		}
		{
			// `Data() ([]byte, error)` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("Data").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Index().Byte()
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.Id("buf").Op(":=").New(Qual("bytes", "Buffer"))
					body.If(
						Err().Op(":=").Qual("github.com/dfuse-io/binary", "NewEncoder").Call(Id("buf")).Dot("Encode").Call(Id("i")).
							Op(";").
							Err().Op("!=").Nil(),
					).Block(
						Return(List(Nil(), Qual("fmt", "Errorf").Call(Lit("unable to encode instruction: %w"), Err()))),
					)
					body.Return(Id("buf").Dot("Bytes").Call(), Nil())
				})
			file.Add(code.Line())
		}
		{
			// `TextEncode(encoder *text.Encoder, option *text.Option) error` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("TextEncode").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("encoder").Op("*").Qual("github.com/gagliardetto/solana-go/text", "Encoder")
						params.Id("option").Op("*").Qual("github.com/gagliardetto/solana-go/text", "Option")
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.Return(Id("encoder").Dot("Encode").Call(Id("inst").Dot("Impl"), Id("option")))
				})
			file.Add(code.Line())
		}
		{
			// `UnmarshalBinary(decoder *bin.Decoder) error` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("UnmarshalBinary").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("decoder").Op("*").Qual("github.com/dfuse-io/binary", "Decoder")
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.Return(Id("inst").Dot("BaseVariant").Dot("UnmarshalBinaryVariant").Call(Id("decoder"), Id("InstructionImplDef")))
				})
			file.Add(code.Line())
		}
		{
			// `MarshalBinary(encoder *bin.Encoder) error ` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("MarshalBinary").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("encoder").Op("*").Qual("github.com/dfuse-io/binary", "Encoder")
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.Err().Op(":=").Id("encoder").Dot("WriteUint32").Call(Id("inst").Dot("TypeID"), Qual("encoding/binary", "LittleEndian"))

					body.If(
						Err().Op("!=").Nil(),
					).Block(
						Return(List(Nil(), Qual("fmt", "Errorf").Call(Lit("unable to write variant type: %w"), Err()))),
					)
					body.Return(Id("encoder").Dot("Encode").Call(Id("inst").Dot("Impl")))
				})
			file.Add(code.Line())
		}
		{
			// `registryDecodeInstruction` func:
			code := Empty()
			code.Func().Id("registryDecodeInstruction").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("accounts").Index().Op("*").Qual("github.com/gagliardetto/solana-go", "AccountMeta")
						params.Id("data").Index().Byte()
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Interface()
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.List(Id("inst"), Err()).Op(":=").Id("DecodeInstruction").Call(Id("accounts"), Id("data"))

					body.If(
						Err().Op("!=").Nil(),
					).Block(
						Return(Nil(), Err()),
					)
					body.Return(Id("inst"), Nil())
				})
			file.Add(code.Line())
		}
		{
			// `DecodeInstruction` func:
			code := Empty()
			code.Func().Id("DecodeInstruction").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("accounts").Index().Op("*").Qual("github.com/gagliardetto/solana-go", "AccountMeta")
						params.Id("data").Index().Byte()
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Op("*").Id("Instruction")
						results.Error()
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:

					body.Id("inst").Op(":=").New(Id("Instruction"))

					body.If(
						Err().Op(":=").Qual("github.com/dfuse-io/binary", "NewDecoder").Call(Id("data")).Dot("Decode").Call(Id("inst")).
							Op(";").
							Err().Op("!=").Nil(),
					).Block(
						Return(
							Nil(),
							Qual("fmt", "Errorf").Call(Lit("unable to decode instruction: %w"), Err()),
						),
					)

					body.If(

						List(Id("v"), Id("ok")).Op(":=").Id("inst").Dot("Impl").Op(".").Parens(Qual("github.com/gagliardetto/solana-go", "AccountsSettable")).
							Op(";").
							Id("ok"),
					).BlockFunc(func(gr *Group) {
						gr.Err().Op(":=").Id("v").Dot("SetAccounts").Call(Id("accounts"))
						gr.If(Nil().Op("!=").Nil()).Block(
							Return(
								Nil(),
								Qual("fmt", "Errorf").Call(Lit("unable to set accounts for instruction: %w"), Err()),
							),
						)
					})

					body.Return(Id("inst"), Nil())
				})
			file.Add(code.Line())
		}
	}

	// Instructions:
	for _, instruction := range idl.Instructions {
		insExportedName := ToCamel(instruction.Name)

		fmt.Println(RedBG(instruction.Name))

		{
			code := Empty().Line().Line()

			for _, doc := range instruction.Docs {
				code.Comment(doc).Line()
			}

			if len(instruction.Docs) == 0 {
				code.Commentf(
					"%s is the `%s` instruction.",
					insExportedName,
					instruction.Name,
				).Line()
			}

			code.Type().Id(insExportedName).StructFunc(func(fieldsGroup *Group) {
				for _, arg := range instruction.Args {
					fieldsGroup.Id(ToCamel(arg.Name)).Add(
						DoGroup(func(fieldTypeGroup *Group) {
							setFieldNameAndType(fieldTypeGroup, arg)
						}),
					)
				}

				fieldsGroup.Line()

				fieldsGroup.Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice").Tag(map[string]string{
					"bin": "-",
				})
			})

			file.Add(code.Line())
		}

		if len(instruction.Accounts) > 0 {
			builderFuncName := "New" + insExportedName + "InstructionBuilder"
			code := Empty()
			code.Commentf(
				"%s creates a new `%s` instruction builder.",
				builderFuncName,
				insExportedName,
			).Line()
			//
			code.Func().Id(builderFuncName).Params().Op("*").Id(insExportedName).
				BlockFunc(func(gr *Group) {
					gr.Return().Op("&").Id(insExportedName).Block(
						Id("AccountMetaSlice").Op(":").Make(Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice"), Lit(instruction.Accounts.NumAccounts())).Op(","),
					)
				})
			file.Add(code.Line())
		}

		{
			// Create parameters setters:
			code := Empty()
			for _, arg := range instruction.Args {
				exportedArgName := ToCamel(arg.Name)

				code.Line().Line()
				for _, doc := range arg.Docs {
					code.Comment(doc).Line()
				}

				code.Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Set" + exportedArgName).
					Params(
						ListFunc(func(st *Group) {
							// Parameters:
							st.Id(arg.Name).Add(idlTypeToType(arg.Type))
							// TODO: determine the right type for the arg.

						}),
					).
					Params(
						ListFunc(func(st *Group) {
							// Results:
							st.Op("*").Id(insExportedName)
						}),
					).
					BlockFunc(func(gr *Group) {
						// Body:
						gr.Id("inst").Dot(exportedArgName).Op("=").Id(arg.Name)

						gr.Return().Id("inst")
					})
			}

			file.Add(code.Line())
		}
		{
			// Account setters/getters:
			code := Empty()
			index := -1
			for _, account := range instruction.Accounts {
				spew.Dump(account)
				// single account (???)
				// TODO: is this a parameter, or a hardcoded value?
				if account.IdlAccount != nil {
					index++
					exportedAccountName := ToCamel(account.IdlAccount.Name)
					lowerAccountName := ToLowerCamel(account.IdlAccount.Name)

					code.Add(createAccountGetterSetter(
						insExportedName,
						account.IdlAccount,
						index,
						exportedAccountName,
						lowerAccountName,
					))
				}

				// many accounts (???)
				// TODO: are these all the wanted parameter accounts, or a list of valid accounts?
				if account.IdlAccounts != nil {
					// builder struct for this accounts group:
					builderStructName := insExportedName + ToCamel(account.IdlAccounts.Name) + "AccountsBuilder"
					code.Line().Line().Type().Id(builderStructName).Struct(
						Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice").Tag(map[string]string{
							"bin": "-",
						}),
					)

					// func that returns a new builder for this account group:
					code.Line().Line().Func().Id("New" + builderStructName).Params().Op("*").Id(builderStructName).
						BlockFunc(func(gr *Group) {
							gr.Return().Op("&").Id(builderStructName).Block(
								Id("AccountMetaSlice").Op(":").Make(Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice"), Lit(account.IdlAccounts.Accounts.NumAccounts())).Op(","),
							)
						}).Line().Line()

					// Method on intruction builder that accepts the accounts group builder, and copies the accounts:
					code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Set" + ToCamel(account.IdlAccounts.Name) + "AccountsFromBuilder").
						Params(
							ListFunc(func(st *Group) {
								// Parameters:
								st.Id(ToLowerCamel(builderStructName)).Op("*").Id(builderStructName)
							}),
						).
						Params(
							ListFunc(func(st *Group) {
								// Results:
								st.Op("*").Id(insExportedName)
							}),
						).
						BlockFunc(func(gr *Group) {
							// Body:

							tpIndex := index
							for _, subAccount := range account.IdlAccounts.Accounts {
								tpIndex++
								exportedAccountName := ToCamel(subAccount.IdlAccount.Name)

								def := Id("inst").Dot("AccountMetaSlice").Index(Lit(tpIndex)).
									Op("=").Id(ToLowerCamel(builderStructName)).Dot("Get" + exportedAccountName + "Account").Call()

								gr.Add(def)
							}

							gr.Return().Id("inst")
						})

					for _, subAccount := range account.IdlAccounts.Accounts {
						index++
						exportedAccountName := ToCamel(subAccount.IdlAccount.Name)
						lowerAccountName := ToLowerCamel(subAccount.IdlAccount.Name)

						code.Add(createAccountGetterSetter(
							builderStructName,
							subAccount.IdlAccount,
							index,
							exportedAccountName,
							lowerAccountName,
						))
					}
				}

			}

			file.Add(code.Line())
		}
		{
			// Add `Build` method to instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Build").
				Params(
					ListFunc(func(st *Group) {
						// Parameters:
					}),
				).
				Params(
					ListFunc(func(st *Group) {
						// Results:
						st.Op("*").Id("Instruction")
					}),
				).
				BlockFunc(func(gr *Group) {
					// Body:

					gr.Return().Op("&").Id("Instruction").Values(
						Dict{
							Id("BaseVariant"): Qual("github.com/dfuse-io/binary", "BaseVariant").Values(
								Dict{
									Id("TypeID"): Id("Instruction_" + insExportedName),
									Id("Impl"):   Id("inst"),
								},
							),
						},
					)
				})
			file.Add(code.Line())
		}
		{
			// Add `Verify` method to instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Verify").
				Params(
					ListFunc(func(st *Group) {
						// Parameters:
					}),
				).
				Params(
					ListFunc(func(st *Group) {
						// Results:
						st.Error()
					}),
				).
				BlockFunc(func(gr *Group) {
					// Body:

					gr.For(List(Id("accIndex"), Id("acc")).Op(":=").Range().Id("inst").Dot("AccountMetaSlice")).Block(
						If(Id("acc").Op("==").Nil()).Block(
							Return(Qual("fmt", "Errorf").Call(List(Lit("ins.AccountMetaSlice[%v] is nil"), Id("accIndex")))),
						),
					)

					gr.Return(Nil())
				})
			file.Add(code.Line())
		}
	}

	{
		// Types:
		for _, typ := range idl.Types {
			switch typ.Type.Kind {
			case IdlTypeDefTyKindStruct:
				code := Empty()
				code.Type().Id(typ.Name).StructFunc(func(fieldsGroup *Group) {
					for _, field := range *typ.Type.Fields {
						fieldsGroup.Id(ToCamel(field.Name)).Add(
							DoGroup(func(fieldTypeGroup *Group) {
								setFieldNameAndType(fieldTypeGroup, field)
							}),
						)
					}
				})

				file.Add(code.Line())
			case IdlTypeDefTyKindEnum:
				code := Empty()
				enumTypeName := typ.Name
				code.Type().Id(enumTypeName).String()

				code.Line().Const().Parens(DoGroup(func(gr *Group) {
					for _, variant := range typ.Type.Variants {
						gr.Id(variant.Name).Id(enumTypeName).Op("=").Lit(variant.Name).Line()
					}
					// TODO: check for fields, etc.
				}))
				file.Add(code.Line())

				// panic(Sf("not implemented: %s", spew.Sdump(typ)))
			default:
				panic(Sf("not implemented: %s", spew.Sdump(typ.Type.Kind)))
			}
		}
	}

	{
		// Accounts:
		for _, acc := range idl.Accounts {
			switch acc.Type.Kind {
			case IdlTypeDefTyKindStruct:
				code := Empty()
				code.Type().Id(acc.Name).StructFunc(func(fieldsGroup *Group) {
					for _, field := range *acc.Type.Fields {
						fieldsGroup.Id(ToCamel(field.Name)).Add(
							DoGroup(func(fieldTypeGroup *Group) {
								setFieldNameAndType(fieldTypeGroup, field)
							}),
						)
					}
				})

				file.Add(code.Line())
			case IdlTypeDefTyKindEnum:
				panic("not implemented")
			default:
				panic("not implemented")
			}
		}
	}

	{
		err := file.Render(os.Stdout)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func createAccountGetterSetter(
	receiverTypeName string,
	account *IdlAccount,
	index int,
	exportedAccountName string,
	lowerAccountName string,
) Code {
	code := Empty().Line().Line()

	for _, doc := range account.Docs {
		code.Comment(doc).Line()
	}
	// Create account setters:
	code.Func().Params(Id("inst").Op("*").Id(receiverTypeName)).Id("Set" + exportedAccountName + "Account").
		Params(
			ListFunc(func(st *Group) {
				// Parameters:
				st.Id(lowerAccountName).Qual("github.com/gagliardetto/solana-go", "PublicKey")
			}),
		).
		Params(
			ListFunc(func(st *Group) {
				// Results:
				st.Op("*").Id(receiverTypeName)
			}),
		).
		BlockFunc(func(gr *Group) {
			// Body:
			def := Id("inst").Dot("AccountMetaSlice").Index(Lit(index)).
				Op("=").Qual("github.com/gagliardetto/solana-go", "NewMeta").Call(Id(lowerAccountName))
			if account.IsMut {
				def.Dot("WRITE").Call()
			}
			if account.IsSigner {
				def.Dot("SIGNER").Call()
			}

			gr.Add(def)

			gr.Return().Id("inst")
		})

	// Create account getters:
	code.Line().Line().Func().Params(Id("inst").Op("*").Id(receiverTypeName)).Id("Get" + exportedAccountName + "Account").
		Params(
			ListFunc(func(st *Group) {
				// Parameters:
			}),
		).
		Params(
			ListFunc(func(st *Group) {
				// Results:
				st.Op("*").Qual("github.com/gagliardetto/solana-go", "AccountMeta")
			}),
		).
		BlockFunc(func(gr *Group) {
			// Body:
			gr.Return(Id("inst").Dot("AccountMetaSlice").Index(Lit(index)))
		})

	return code
}

func setFieldNameAndType(fieldTypeGroup *Group, idlField IdlField) {
	if idlField.Type.IsString() {
		fieldTypeGroup.Add(typeStringToType(idlField.Type.GetString()))
	} else if idlField.Type.IsIdlTypeDefined() {
		fieldTypeGroup.Add(Id(idlField.Type.GetIdlTypeDefined().Defined))
	} else if idlField.Type.IsArray() {
		arr := idlField.Type.GetArray()

		if arr.Thing.IsString() {
			fieldTypeGroup.Index()
			fieldTypeGroup.Add(typeStringToType(arr.Thing.GetString()))
		} else if arr.Thing.IsIdlTypeDefined() {
			fieldTypeGroup.Index()
			fieldTypeGroup.Add(Id(arr.Thing.GetIdlTypeDefined().Defined))
		} else {
			panic(spew.Sdump(arr))
		}
	} else {
		panic(spew.Sdump(idlField))
	}
}
