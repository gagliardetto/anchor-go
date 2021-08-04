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
		// Base Instruction struct:
		code := Empty()
		code.Type().Id("Instruction").Struct(
			Qual("github.com/dfuse-io/binary", "BaseVariant"),
		)

		file.Add(code.Line())
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
							setType(fieldTypeGroup, arg)
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

				code.Func().Params(Id("ins").Op("*").Id(insExportedName)).Id("Set" + exportedArgName).
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
						gr.Id("ins").Dot(exportedArgName).Op("=").Id(arg.Name)

						gr.Return().Id("ins")
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
					builderStructName := ToCamel(account.IdlAccounts.Name) + "AccountsBuilder"
					code.Line().Line().Type().Id(builderStructName).Struct(
						Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice").Tag(map[string]string{
							"bin": "-",
						}),
					)

					// func that returns a new builder for this account group:
					code.Line().Line().Func().Id("New" + builderStructName).Params().Op("*").Id(builderStructName).
						BlockFunc(func(gr *Group) {
							gr.Return().Op("&").Id(insExportedName).Block(
								Id("AccountMetaSlice").Op(":").Make(Qual("github.com/gagliardetto/solana-go", "AccountMetaSlice"), Lit(account.IdlAccounts.Accounts.NumAccounts())).Op(","),
							)
						}).Line().Line()

					// MEthod on intruction builder that accepts the accounts group builder, and copies the accounts:
					code.Line().Line().Func().Params(Id("ins").Op("*").Id(insExportedName)).Id("Set" + ToCamel(account.IdlAccounts.Name) + "AccountsFromBuilder").
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

								def := Id("ins").Dot("AccountMetaSlice").Index(Lit(tpIndex)).
									Op("=").Id(ToLowerCamel(builderStructName)).Dot("Get" + exportedAccountName + "Account").Call()

								gr.Add(def)
							}

							gr.Return().Id("ins")
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

			code.Line().Line().Func().Params(Id("ins").Op("*").Id(insExportedName)).Id("Build").
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
									Id("Impl"):   Id("ins"),
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

			code.Line().Line().Func().Params(Id("ins").Op("*").Id(insExportedName)).Id("Verify").
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

					gr.For(List(Id("accIndex"), Id("acc")).Op(":=").Range().Id("ins").Dot("AccountMetaSlice")).Block(
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
								setType(fieldTypeGroup, field)
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
								setType(fieldTypeGroup, field)
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
	code.Func().Params(Id("ins").Op("*").Id(receiverTypeName)).Id("Set" + exportedAccountName + "Account").
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
			def := Id("ins").Dot("AccountMetaSlice").Index(Lit(index)).
				Op("=").Qual("github.com/gagliardetto/solana-go", "NewMeta").Call(Id(lowerAccountName))
			if account.IsMut {
				def.Dot("WRITE").Call()
			}
			if account.IsSigner {
				def.Dot("SIGNER").Call()
			}

			gr.Add(def)

			gr.Return().Id("ins")
		})

	// Create account getters:
	code.Line().Line().Func().Params(Id("ins").Op("*").Id(receiverTypeName)).Id("Get" + exportedAccountName + "Account").
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
			gr.Return(Id("ins").Dot("AccountMetaSlice").Index(Lit(index)))
		})

	return code
}

func setType(fieldTypeGroup *Group, idlField IdlField) {
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
