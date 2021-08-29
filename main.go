package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	. "github.com/dave/jennifer/jen"
	. "github.com/gagliardetto/utilz"
)

func main() {
	// TODO: load config from flags, etc.
	conf.Encoding = EncodingBorsh

	flag.BoolVar(&conf.Debug, "debug", false, "debug mode")
	flag.StringVar((*string)(&conf.Encoding), "codec", "borsh", "Choose codec")
	filenames := FlagStringArray{}
	flag.Var(&filenames, "src", "Path to source; can use multiple times.")
	flag.Parse()

	if err := conf.Validate(); err != nil {
		panic(fmt.Errorf("error while validating config: %w", err))
	}

	filenamesExtra := []string{
		// "idl/swap_light.json",
		// "solana/native/system.json",
		//
		"idl/testing/enum.json",
		// "idl/testing/deeply-nested-accounts.json",

		// "idl/registry.json",
		// "idl/cashiers_check.json",
		// "idl/chat.json",
		// "idl/composite.json",
		// "idl/counter_auth.json",
		// "idl/counter.json",
		// "idl/errors.json",
		// "idl/escrow.json",
		// "idl/events.json",
		// "idl/ido_pool.json",
		// "idl/lockup.json",
		// "idl/misc.json",
		// "idl/multisig.json",
		// "idl/pyth.json",
		// "idl/swap.json",
		// "idl/swap_light.json",
		// "idl/sysvars.json",
		// "idl/typescript.json",
		// "idl/zero_copy.json",
	}
	_ = filenamesExtra

	var ts time.Time
	if GetConfig().Debug {
		ts = time.Unix(0, 0)
	} else {
		ts = time.Now()
	}
	outDir := "generated"

	for _, idlFilepath := range filenames {
		Ln("Generating client for", LimeBG(idlFilepath))
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

		// spew.Dump(idl)

		// Create subfolder for package for generated assets:
		packageAssetFolderName := ToLowerCamel(idl.Name)
		packageAssetFolderPath := path.Join(outDir, packageAssetFolderName)
		MustCreateFolderIfNotExists(packageAssetFolderPath, os.ModePerm)
		// Create folder for assets generated during this run:
		thisRunAssetFolderName := ToLowerCamel(idl.Name) + "_" + ts.Format(FilenameTimeFormat)
		thisRunAssetFolderPath := path.Join(packageAssetFolderPath, thisRunAssetFolderName)
		// Create a new assets folder inside the main assets folder:
		MustCreateFolderIfNotExists(thisRunAssetFolderPath, os.ModePerm)

		files, err := GenerateClientFromProgramIDL(idl)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			// err := file.Render(os.Stdout)
			// if err != nil {
			// 	panic(err)
			// }

			{
				// Save assets:
				assetFileName := file.Name + ".go"
				assetFilepath := path.Join(thisRunAssetFolderPath, assetFileName)

				// Create file:
				goFile, err := os.Create(assetFilepath)
				if err != nil {
					panic(err)
				}
				defer goFile.Close()

				// Write generated code file:
				Infof("Saving assets to %q", MustAbs(assetFilepath))
				err = file.File.Render(goFile)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func GenerateClientFromProgramIDL(idl IDL) ([]*FileWrapper, error) {
	if err := idl.Validate(); err != nil {
		return nil, err
	}

	files := make([]*FileWrapper, 0)
	{
		// Create and populate Go file that holds all the basic
		// elements of an instruction client:
		file, err := genProgramBoilerplate(idl)
		if err != nil {
			return nil, err
		}
		files = append(files, &FileWrapper{
			Name: "instructions",
			File: file,
		})
	}

	{
		file := NewGoFile(idl.Name, true)
		// Declare types from IDL:
		for _, typ := range idl.Types {
			file.Add(genTypeDef(typ))
		}
		files = append(files, &FileWrapper{
			Name: "types",
			File: file,
		})
	}

	{
		file := NewGoFile(idl.Name, true)
		// Declare account layouts from IDL:
		for _, acc := range idl.Accounts {
			file.Add(genTypeDef(acc))
		}
		files = append(files, &FileWrapper{
			Name: "accounts",
			File: file,
		})
	}

	// Instructions:
	for _, instruction := range idl.Instructions {
		file := NewGoFile(idl.Name, true)
		insExportedName := ToCamel(instruction.Name)

		// fmt.Println(RedBG(instruction.Name))

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
				for argIndex, arg := range instruction.Args {
					if len(arg.Docs) > 0 {
						if argIndex > 0 {
							fieldsGroup.Line()
						}
						for _, doc := range arg.Docs {
							fieldsGroup.Comment(doc)
						}
					}
					fieldsGroup.Add(genField(arg, true)).Add(func() Code {
						if arg.Type.IsIdlTypeOption() {
							return Tag(map[string]string{
								"bin": "optional",
							}).Comment("OPTIONAL")
						}
						return nil
					}())
				}

				fieldsGroup.Line()

				{
					lastGroupName := ""
					// Add comments of the accounts from rust docs.
					instruction.Accounts.Walk("", nil, nil, func(groupPath string, accountIndex int, parentGroup *IdlAccounts, ia *IdlAccount) bool {
						comment := &strings.Builder{}
						indent := 6
						var prepend int

						if groupPath != "" {
							thisGroupName := filepath.Base(groupPath)
							indent = len(thisGroupName) + 2
							if strings.Count(groupPath, "/") == 0 {
								prepend = 6
							} else {
								prepend = 6 + (strings.Count(groupPath, "/") * 2) + len(strings.TrimSuffix(groupPath, thisGroupName)) - 1
							}
							if lastGroupName != groupPath {
								comment.WriteString(strings.Repeat("·", prepend-1) + Sf(" %s: ", thisGroupName))
							} else {
								comment.WriteString(strings.Repeat("·", prepend+indent-1) + " ")
							}
							lastGroupName = groupPath
						}

						comment.WriteString(Sf("[%v] = ", accountIndex))
						comment.WriteString("[")
						if ia.IsMut {
							comment.WriteString("WRITE")
						}
						if ia.IsSigner {
							if ia.IsMut {
								comment.WriteString(", ")
							}
							comment.WriteString("SIGNER")
						}
						comment.WriteString("] ")
						comment.WriteString(ia.Name)

						fieldsGroup.Comment(comment.String())
						for _, doc := range ia.Docs {
							fieldsGroup.Comment(strings.Repeat("·", prepend+indent-1+6) + " " + doc)
						}
						if accountIndex < instruction.Accounts.NumAccounts()-1 {
							fieldsGroup.Comment("")
						}

						accountIndex++
						return true
					})
				}
				fieldsGroup.Qual(PkgSolanaGo, "AccountMetaSlice").Tag(map[string]string{
					"bin":        "-",
					"borsh_skip": "true",
				})
			})

			file.Add(code.Line())
		}

		if len(instruction.Accounts) > 0 {
			builderFuncName := formatBuilderFuncName(insExportedName)
			code := Empty()
			code.Commentf(
				"%s creates a new `%s` instruction builder.",
				builderFuncName,
				insExportedName,
			).Line()
			//
			code.Func().Id(builderFuncName).Params().Op("*").Id(insExportedName).
				BlockFunc(func(body *Group) {
					body.Id("nd").Op(":=").Op("&").Id(insExportedName).Block(
						Id("AccountMetaSlice").Op(":").Make(Qual(PkgSolanaGo, "AccountMetaSlice"), Lit(instruction.Accounts.NumAccounts())).Op(","),
					)

					// Set sysvar accounts:
					instruction.Accounts.Walk("", nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, account *IdlAccount) bool {
						if isVar(account.Name) {
							pureVarName := getVarName(account.Name)
							is := isSysVar(pureVarName)
							if is {
								_, ok := sysVars[pureVarName]
								if !ok {
									panic(account)
								}
								def := Qual(PkgSolanaGo, "Meta").Call(Qual(PkgSolanaGo, pureVarName))
								if account.IsMut {
									def.Dot("WRITE").Call()
								}
								if account.IsSigner {
									def.Dot("SIGNER").Call()
								}

								body.Id("nd").Dot("AccountMetaSlice").Index(Lit(index)).Op("=").Add(def)
							} else {
								panic(account)
							}
						}
						return true
					})

					body.Return(Id("nd"))
				})
			file.Add(code.Line())
		}

		{
			// Declare methods that set the parameters of the instruction:
			code := Empty()
			for _, arg := range instruction.Args {
				exportedArgName := ToCamel(arg.Name)

				code.Line().Line()
				for _, doc := range arg.Docs {
					code.Comment(doc).Line()
				}

				code.Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Set" + exportedArgName).
					Params(
						ListFunc(func(params *Group) {
							// Parameters:
							params.Id(arg.Name).Add(genTypeName(arg.Type))
						}),
					).
					Params(
						ListFunc(func(results *Group) {
							// Results:
							results.Op("*").Id(insExportedName)
						}),
					).
					BlockFunc(func(body *Group) {
						// Body:
						body.Id("inst").Dot(exportedArgName).Op("=").
							Add(func() Code {
								if isComplexEnum(arg.Type) {
									return nil
								}
								return Op("&")
							}()).
							Id(arg.Name)

						body.Return().Id("inst")
					})
			}

			file.Add(code.Line())
		}
		{
			// Declare methods that set/get accounts for the instruction:
			code := Empty()

			declaredReceivers := []string{}
			groupMemberIndex := 0
			instruction.Accounts.Walk("", nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, account *IdlAccount) bool {
				builderStructName := insExportedName + ToCamel(parentGroupPath) + "AccountsBuilder"
				hasNestedParent := parentGroupPath != ""
				isDeclaredReceiver := SliceContains(declaredReceivers, parentGroupPath)

				if !hasNestedParent {
					groupMemberIndex = index
				}
				if hasNestedParent && !isDeclaredReceiver {
					groupMemberIndex = 0
					declaredReceivers = append(declaredReceivers, parentGroupPath)
					// many accounts (???)
					// builder struct for this accounts group:

					code.Line().Line()
					for _, doc := range parentGroup.Docs {
						code.Comment(doc).Line()
					}
					code.Type().Id(builderStructName).Struct(
						Qual(PkgSolanaGo, "AccountMetaSlice").Tag(map[string]string{
							"bin":        "-",
							"borsh_skip": "true",
						}),
					)

					// func that returns a new builder for this account group:
					code.Line().Line().Func().Id("New" + builderStructName).Params().Op("*").Id(builderStructName).
						BlockFunc(func(gr *Group) {
							gr.Return().Op("&").Id(builderStructName).Block(
								Id("AccountMetaSlice").Op(":").Make(Qual(PkgSolanaGo, "AccountMetaSlice"), Lit(parentGroup.Accounts.NumAccounts())).Op(","),
							)
						}).Line().Line()

					// Method on intruction builder that accepts the accounts group builder, and copies the accounts:
					code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Set" + ToCamel(parentGroup.Name) + "AccountsFromBuilder").
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
							// spew.Dump(parentGroup)
							for _, subAccount := range parentGroup.Accounts {
								if subAccount.IdlAccount != nil {
									exportedAccountName := ToCamel(subAccount.IdlAccount.Name)

									def := Id("inst").Dot("AccountMetaSlice").Index(Lit(tpIndex)).
										Op("=").Id(ToLowerCamel(builderStructName)).Dot("Get" + exportedAccountName + "Account").Call()

									gr.Add(def)
								}
								tpIndex++
							}

							gr.Return().Id("inst")
						})
				}

				{
					exportedAccountName := ToCamel(account.Name)
					lowerAccountName := ToLowerCamel(account.Name)

					var receiverTypeName string
					if parentGroupPath == "" {
						receiverTypeName = insExportedName
					} else {
						receiverTypeName = builderStructName
					}

					code.Add(genAccountGettersSetters(
						receiverTypeName,
						account,
						groupMemberIndex,
						exportedAccountName,
						lowerAccountName,
					))
					groupMemberIndex++
				}
				return true
			})

			file.Add(code.Line())
		}
		{
			// Declare `Build` method on instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Id(insExportedName)).Id("Build").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Op("*").Id("Instruction")
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:

					body.Return().Op("&").Id("Instruction").Values(
						Dict{
							Id("BaseVariant"): Qual(PkgDfuseBinary, "BaseVariant").Values(
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
			// Declare `Validate` method on instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("Validate").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
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
					if len(instruction.Args) > 0 {
						body.Comment("Check whether all (required) parameters are set:")

						body.BlockFunc(func(paramVerifyBody *Group) {
							for _, arg := range instruction.Args {
								exportedArgName := ToCamel(arg.Name)

								// Optional params can be empty.
								if arg.Type.IsIdlTypeOption() {
									continue
								}

								paramVerifyBody.If(Id("inst").Dot(exportedArgName).Op("==").Nil()).Block(
									Return(
										Qual("errors", "New").Call(Lit(Sf("%s parameter is not set", exportedArgName))),
									),
								)
							}
						})
						body.Line()
					}

					body.Comment("Check whether all accounts are set:")
					body.For(List(Id("accIndex"), Id("acc")).Op(":=").Range().Id("inst").Dot("AccountMetaSlice")).Block(
						If(Id("acc").Op("==").Nil()).Block(
							Return(Qual("fmt", "Errorf").Call(List(Lit("ins.AccountMetaSlice[%v] is not set"), Id("accIndex")))),
						),
					)

					body.Return(Nil())
				})
			file.Add(code.Line())
		}
		{
			// Declare `EncodeToTree(parent treeout.Branches)` method in instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("EncodeToTree").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("parent").Qual(PkgTreeout, "Branches")
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:

					body.Id("parent").Dot("Child").Call(Qual(PkgFormat, "Program").Call(Id("ProgramName"), Id("ProgramID"))).Op(".").
						Line().Comment("").Line().
						Id("ParentFunc").Parens(Func().Parens(Id("programBranch").Qual(PkgTreeout, "Branches")).BlockFunc(
						func(programBranchGroup *Group) {
							programBranchGroup.Id("programBranch").Dot("Child").Call(Qual(PkgFormat, "Instruction").Call(Lit(insExportedName))).Op(".").
								Line().Comment("").Line().
								Id("ParentFunc").Parens(Func().Parens(Id("instructionBranch").Qual(PkgTreeout, "Branches")).BlockFunc(
								func(instructionBranchGroup *Group) {

									instructionBranchGroup.Line().Comment("Parameters of the instruction:")

									instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit("Params")).Dot("ParentFunc").Parens(Func().Parens(Id("paramsBranch").Qual(PkgTreeout, "Branches")).BlockFunc(func(paramsBranchGroup *Group) {
										for _, arg := range instruction.Args {
											exportedArgName := ToCamel(arg.Name)
											paramsBranchGroup.Id("paramsBranch").Dot("Child").
												Call(
													Qual(PkgFormat, "Param").Call(
														Lit(exportedArgName+StringIf(arg.Type.IsIdlTypeOption(), " (OPTIONAL)")),
														Add(CodeIf(!arg.Type.IsIdlTypeOption(), Op("*"))).Id("inst").Dot(exportedArgName),
													),
												)
										}
									}))

									instructionBranchGroup.Line().Comment("Accounts of the instruction:")

									instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit("Accounts")).Dot("ParentFunc").Parens(
										Func().Parens(Id("accountsBranch").Qual(PkgTreeout, "Branches")).BlockFunc(func(accountsBranchGroup *Group) {

											instruction.Accounts.Walk("", nil, nil, func(groupPath string, accountIndex int, parentGroup *IdlAccounts, ia *IdlAccount) bool {
												exportedAccountName := filepath.Join(groupPath, ia.Name)
												accountsBranchGroup.Id("accountsBranch").Dot("Child").Call(Qual(PkgFormat, "Meta").Call(Lit(exportedAccountName), Id("inst").Dot("AccountMetaSlice").Index(Lit(accountIndex))))
												return true
											})
										}))
								}))
						}))
				})
			file.Add(code.Line())
		}

		{
			// Declare `MarshalWithEncoder(encoder *bin.Encoder) error` method on instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Id(insExportedName)).Id("MarshalWithEncoder").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("encoder").Op("*").Qual(PkgDfuseBinary, "Encoder")
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
					for _, arg := range instruction.Args {
						exportedArgName := ToCamel(arg.Name)
						body.Commentf("Serialize `%s` param:", exportedArgName)

						if isComplexEnum(arg.Type) {
							enumName := arg.Type.GetIdlTypeDefined().Defined
							body.BlockFunc(func(argBody *Group) {
								argBody.List(Id("tmp")).Op(":=").Id(formatEnumContainerName(enumName)).Block()
								argBody.Switch(Id("realvalue").Op(":=").Id("inst").Dot(exportedArgName).Op(".").Parens(Type())).
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
							body.BlockFunc(func(argBody *Group) {
								// TODO: check if is optional and nil
								argBody.Err().Op(":=").Id("encoder").Dot("Encode").Call(Add(CodeIf(!arg.Type.IsIdlTypeOption(), Op("*"))).Id("inst").Dot(exportedArgName))

								argBody.If(
									Err().Op("!=").Nil(),
								).Block(
									Return(Err()),
								)

							})
						}

					}

					body.Return(Nil())
				})
			file.Add(code.Line())
		}

		{
			// Declare `UnmarshalWithDecoder(decoder *bin.Decoder) error` method on instruction:
			code := Empty()

			code.Line().Line().Func().Params(Id("inst").Op("*").Id(insExportedName)).Id("UnmarshalWithDecoder").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("decoder").Op("*").Qual(PkgDfuseBinary, "Decoder")
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
					for _, arg := range instruction.Args {
						exportedArgName := ToCamel(arg.Name)
						body.Commentf("Deserialize `%s` param:", exportedArgName)

						if isComplexEnum(arg.Type) {
							enumName := arg.Type.GetIdlTypeDefined().Defined
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
											switchGroup.Case(Lit(variantIndex)).
												BlockFunc(func(caseGroup *Group) {
													caseGroup.Id("inst").Dot(exportedArgName).Op("=").Op("&").Id("tmp").Dot(ToCamel(variant.Name))
												})
										}
										switchGroup.Default().
											BlockFunc(func(caseGroup *Group) {
												caseGroup.Return(Qual("fmt", "Errorf").Call(Lit("unknown enum index: %v"), Id("tmp").Dot("Enum")))
											})
									})

							})
						} else {
							body.BlockFunc(func(argBody *Group) {
								argBody.Err().Op(":=").Id("decoder").Dot("Decode").Call(Op("&").Id("inst").Dot(exportedArgName))

								argBody.If(
									Err().Op("!=").Nil(),
								).Block(
									Return(Err()),
								)
							})
						}

					}

					body.Return(Nil())
				})
			file.Add(code.Line())
		}

		{
			// Declare instruction initializer func:
			code := Empty()
			code.Func().Id("New" + insExportedName).
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						{

							for argIndex, arg := range instruction.Args {
								params.Add(func() Code {
									if argIndex == 0 {
										return Line().Comment("Parameters:")
									}
									return Empty()
								}()).Line().Id(arg.Name).Add(genTypeName(arg.Type))
							}
						}
						{
							instruction.Accounts.Walk("", nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, account *IdlAccount) bool {
								// skip sysvars:
								if isSysVar(account.Name) {
									return true
								}
								var accountName string
								if parentGroupPath == "" {
									accountName = ToLowerCamel(account.Name)
								} else {
									accountName = ToLowerCamel(parentGroupPath + "/" + ToLowerCamel(account.Name))
								}

								params.Add(func() Code {
									if index == 0 {
										return Line().Comment("Accounts:").Line()
									}
									return Line()
								}()).Id(accountName).Qual(PkgSolanaGo, "PublicKey")
								return true
							})
						}
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
						results.Op("*").Id("Instruction")
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					builder := body.Return().Id(formatBuilderFuncName(insExportedName)).Call()
					{
						for _, arg := range instruction.Args {
							exportedArgName := ToCamel(arg.Name)
							builder.Op(".").Line().Id("Set" + exportedArgName).Call(Id(arg.Name))
						}
					}

					{
						declaredReceivers := []string{}
						instruction.Accounts.Walk("", nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, account *IdlAccount) bool {
							// skip sysvars:
							if isSysVar(account.Name) {
								return true
							}
							var accountName string
							if parentGroupPath == "" {
								accountName = ToLowerCamel(account.Name)
							} else {
							}

							builderStructName := insExportedName + ToCamel(parentGroupPath) + "AccountsBuilder"
							hasNestedParent := parentGroupPath != ""
							isDeclaredReceiver := SliceContains(declaredReceivers, parentGroupPath)

							if hasNestedParent && !isDeclaredReceiver {
								declaredReceivers = append(declaredReceivers, parentGroupPath)
								builder.Op(".").Line().Id("Set" + ToCamel(parentGroup.Name) + "AccountsFromBuilder").Call(
									Line().Id("New" + builderStructName).Call().
										Add(
											DoGroup(func(gr *Group) {
												// Body:
												for subIndex, subAccount := range parentGroup.Accounts {
													if subAccount.IdlAccount != nil {
														exportedAccountName := ToCamel(subAccount.IdlAccount.Name)
														accountName = ToLowerCamel(parentGroupPath + "/" + ToLowerCamel(exportedAccountName))

														gr.Op(".").Add(func() Code {
															if subIndex == 0 {
																return Line().Line()
															}
															return Line()
														}()).Id("Set" + exportedAccountName + "Account").Call(Id(accountName))

														if subIndex == len(parentGroup.Accounts)-1 {
															gr.Op(",").Line()
														}
													}
												}
											}),
										),
								)
							}

							if !hasNestedParent {
								builder.Op(".").Line().Id("Set" + ToCamel(account.Name) + "Account").Call(Id(accountName))
							}

							return true
						})
					}

					builder.Op(".").Line().Id("Build").Call()
				})

			file.Add(code.Line())
		}
		////
		files = append(files, &FileWrapper{
			Name: insExportedName,
			File: file,
		})
	}

	{
		testFiles, err := genTestingFuncs(idl)
		if err != nil {
			return nil, err
		}
		files = append(files, testFiles...)
	}

	return files, nil
}

func genAccountGettersSetters(
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
			ListFunc(func(params *Group) {
				// Parameters:
				params.Id(lowerAccountName).Qual(PkgSolanaGo, "PublicKey")
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				// Results:
				results.Op("*").Id(receiverTypeName)
			}),
		).
		BlockFunc(func(body *Group) {
			// Body:
			def := Id("inst").Dot("AccountMetaSlice").Index(Lit(index)).
				Op("=").Qual(PkgSolanaGo, "Meta").Call(Id(lowerAccountName))
			if account.IsMut {
				def.Dot("WRITE").Call()
			}
			if account.IsSigner {
				def.Dot("SIGNER").Call()
			}

			body.Add(def)

			body.Return().Id("inst")
		})

	// Create account getters:
	code.Line().Line().Func().Params(Id("inst").Op("*").Id(receiverTypeName)).Id("Get" + exportedAccountName + "Account").
		Params(
			ListFunc(func(params *Group) {
				// Parameters:
			}),
		).
		Params(
			ListFunc(func(results *Group) {
				// Results:
				results.Op("*").Qual(PkgSolanaGo, "AccountMeta")
			}),
		).
		BlockFunc(func(body *Group) {
			// Body:
			body.Return(Id("inst").Dot("AccountMetaSlice").Index(Lit(index)))
		})

	return code
}

func genProgramBoilerplate(idl IDL) (*File, error) {
	file := NewGoFile(idl.Name, true)
	for _, programDoc := range idl.Docs {
		file.HeaderComment(programDoc)
	}

	{
		// ProgramID variable:
		code := Empty()

		hasAddress := idl.Metadata != nil && idl.Metadata.Address != ""
		code.Var().Id("ProgramID").Qual(PkgSolanaGo, "PublicKey").
			Add(
				func() Code {
					if hasAddress {
						return Op("=").Qual(PkgSolanaGo, "MustPublicKeyFromBase58").Call(Lit(idl.Metadata.Address))
					}
					return nil
				}(),
			)
		file.Add(code.Line())
	}
	{
		// `SetProgramID` func:
		code := Empty()
		code.Func().Id("SetProgramID").Params(Id("pubkey").Qual(PkgSolanaGo, "PublicKey")).Block(
			Id("ProgramID").Op("=").Id("pubkey"),
			Qual(PkgSolanaGo, "RegisterInstructionDecoder").Call(Id("ProgramID"), Id("registryDecodeInstruction")),
		)
		file.Add(code.Line())
	}
	{
		// ProgramName variable:
		code := Empty()
		programName := ToCamel(idl.Name)
		code.Const().Id("ProgramName").Op("=").Lit(programName)
		file.Add(code.Line())
	}
	{
		// register decoder:
		code := Empty()
		code.Func().Id("init").Call().Block(
			If(
				Op("!").Id("ProgramID").Dot("IsZero").Call(),
			).Block(
				Qual(PkgSolanaGo, "RegisterInstructionDecoder").Call(Id("ProgramID"), Id("registryDecodeInstruction")),
			),
		)
		file.Add(code.Line())
	}

	{
		// Instruction ID enum:
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
		// InstructionIDToName
		code := Empty()
		code.Comment("InstructionIDToName returns the name of the instruction given its ID.").Line()
		code.Func().Id("InstructionIDToName").
			Params(Id("id").Uint32()).
			Params(String()).
			BlockFunc(func(body *Group) {
				body.Switch(Id("id")).BlockFunc(func(switchBlock *Group) {
					for _, instruction := range idl.Instructions {
						insExportedName := ToCamel(instruction.Name)
						switchBlock.Case(Id("Instruction_" + insExportedName)).Line().Return(Lit(insExportedName))
					}
					switchBlock.Default().Line().Return(Lit(""))
				})

			})
		file.Add(code.Line())
	}

	{
		{ // Base Instruction struct:
			code := Empty()
			code.Type().Id("Instruction").Struct(
				Qual(PkgDfuseBinary, "BaseVariant"),
			)
			file.Add(code.Line())
		}
		{
			// `EncodeToTree(parent treeout.Branches)` method
			code := Empty()
			code.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("EncodeToTree").
				Params(Id("parent").Qual(PkgTreeout, "Branches")).
				Params().
				BlockFunc(func(body *Group) {
					body.If(
						List(Id("enToTree"), Id("ok")).Op(":=").Id("inst").Dot("Impl").Op(".").Parens(Qual(PkgSolanaGoText, "EncodableToTree")).
							Op(";").
							Id("ok"),
					).Block(
						Id("enToTree").Dot("EncodeToTree").Call(Id("parent")),
					).Else().Block(
						Id("parent").Dot("Child").Call(Qual("github.com/davecgh/go-spew/spew", "Sdump").Call(Id("inst"))),
					)
				})
			file.Add(code.Line())
		}
		{
			// variant definitions for the decoder:
			code := Empty()
			code.Var().Id("InstructionImplDef").Op("=").Qual(PkgDfuseBinary, "NewVariantDefinition").
				Parens(DoGroup(func(call *Group) {
					call.Line()
					// TODO: make this configurable?
					call.Qual(PkgDfuseBinary, "Uint32TypeIDEncoding").Op(",").Line()

					call.Index().Qual(PkgDfuseBinary, "VariantType").
						BlockFunc(func(variantBlock *Group) {
							for _, instruction := range idl.Instructions {
								insName := ToCamel(instruction.Name)
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
				Parens(Qual(PkgSolanaGo, "PublicKey")).
				BlockFunc(func(body *Group) {
					body.Return(
						Id("ProgramID"),
					)
				})
			file.Add(code.Line())
		}
		{
			// method to return accounts:
			code := Empty()
			code.Func().Parens(Id("inst").Op("*").Id("Instruction")).Id("Accounts").Params().
				Parens(Id("out").Index().Op("*").Qual(PkgSolanaGo, "AccountMeta")).
				BlockFunc(func(body *Group) {
					body.Return(
						Id("inst").Dot("Impl").Op(".").Parens(Qual(PkgSolanaGo, "AccountsGettable")).Dot("GetAccounts").Call(),
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
						Err().Op(":=").Qual(PkgDfuseBinary, GetConfig().Encoding._NewEncoder()).Call(Id("buf")).Dot("Encode").Call(Id("inst")).
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
						params.Id("encoder").Op("*").Qual(PkgSolanaGoText, "Encoder")
						params.Id("option").Op("*").Qual(PkgSolanaGoText, "Option")
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
			// `UnmarshalWithDecoder(decoder *bin.Decoder) error` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("UnmarshalWithDecoder").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("decoder").Op("*").Qual(PkgDfuseBinary, "Decoder")
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
			// `MarshalWithEncoder(encoder *bin.Encoder) error ` method:
			code := Empty()
			code.Func().Params(Id("inst").Op("*").Id("Instruction")).Id("MarshalWithEncoder").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("encoder").Op("*").Qual(PkgDfuseBinary, "Encoder")
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
						Return(Qual("fmt", "Errorf").Call(Lit("unable to write variant type: %w"), Err())),
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
						params.Id("accounts").Index().Op("*").Qual(PkgSolanaGo, "AccountMeta")
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
						params.Id("accounts").Index().Op("*").Qual(PkgSolanaGo, "AccountMeta")
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
						Err().Op(":=").Qual(PkgDfuseBinary, GetConfig().Encoding._NewDecoder()).Call(Id("data")).Dot("Decode").Call(Id("inst")).
							Op(";").
							Err().Op("!=").Nil(),
					).Block(
						Return(
							Nil(),
							Qual("fmt", "Errorf").Call(Lit("unable to decode instruction: %w"), Err()),
						),
					)

					body.If(

						List(Id("v"), Id("ok")).Op(":=").Id("inst").Dot("Impl").Op(".").Parens(Qual(PkgSolanaGo, "AccountsSettable")).
							Op(";").
							Id("ok"),
					).BlockFunc(func(gr *Group) {
						gr.Err().Op(":=").Id("v").Dot("SetAccounts").Call(Id("accounts"))
						gr.If(Err().Op("!=").Nil()).Block(
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

	return file, nil
}
