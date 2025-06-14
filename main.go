package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/sighash"
	bin "github.com/gagliardetto/binary"
	. "github.com/gagliardetto/utilz"
	"golang.org/x/mod/modfile"
)

const generatedDir = "generated"

// TODO:
// - tests where type has field that is a complex enum (represented as an interface): assign a random concrete value from the possible enum variants.
// - when printing tree, check for len before accessing array indexes.

func main() {
	conf.Encoding = EncodingBorsh
	conf.TypeID = TypeIDAnchor

	filenames := FlagStringArray{}
	flag.Var(&filenames, "src", "Path to source; can use multiple times.")
	flag.StringVar(&conf.DstDir, "dst", generatedDir, "Destination folder")
	flag.BoolVar(&conf.Debug, "debug", false, "debug mode")
	flag.BoolVar(&conf.RemoveAccountSuffix, "remove-account-suffix", false, "Remove \"Account\" suffix from accessors (if leads to duplication, e.g. \"SetFooAccountAccount\")")

	flag.StringVar((*string)(&conf.Encoding), "codec", string(EncodingBorsh), "Choose codec")
	flag.StringVar((*string)(&conf.TypeID), "type-id", string(TypeIDAnchor), "Choose typeID kind")
	flag.StringVar(&conf.ModPath, "mod", "", "Generate a go.mod file with the necessary dependencies, and this module")
	flag.Parse()

	if err := conf.Validate(); err != nil {
		panic(fmt.Errorf("error while validating config: %w", err))
	}

	var ts time.Time
	if GetConfig().Debug {
		ts = time.Unix(0, 0)
	} else {
		ts = time.Now()
	}
	if len(filenames) == 0 {
		Sfln(
			"[%s] No IDL files provided",
			Red(XMark),
		)
		os.Exit(1)
	}
	{
		exists, err := DirExists(GetConfig().DstDir)
		if err != nil {
			panic(err)
		}
		if !exists {
			if GetConfig().DstDir == generatedDir {
				MustCreateFolderIfNotExists(generatedDir, os.ModePerm)
			} else {
				Sfln(
					"[%s] destination dir doesn't exist: %s",
					Red(XMark),
					GetConfig().DstDir,
				)
				os.Exit(1)
			}
		}
	}

	callbacks := make([]func(), 0)
	defer func() {
		for _, cb := range callbacks {
			cb()
		}
	}()

	for _, idlFilepath := range filenames {
		Sfln(
			"[%s] Generating client from IDL: %s",
			Shakespeare("+"),
			Shakespeare(idlFilepath),
		)
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
		{
			if idl.State != nil {
				Sfln(
					"%s idl.State is defined, but generator is not implemented yet.",
					OrangeBG("[?]"),
				)
			}
			if len(idl.Events) > 0 {
				Sfln(
					"%s idl.Events is defined, but generator is not implemented yet.",
					OrangeBG("[?]"),
				)
			}
			if len(idl.Errors) > 0 {
				Sfln(
					"%s idl.Errors is defined, but generator is not implemented yet.",
					OrangeBG("[?]"),
				)
			}
			if len(idl.Constants) > 0 {
				Sfln(
					"%s idl.Constants is defined, but generator is not implemented yet.",
					OrangeBG("[?]"),
				)
			}
		}

		// spew.Dump(idl)

		// Create subfolder for package for generated assets:
		packageAssetFolderName := sighash.ToRustSnakeCase(idl.Name)
		var dstDirForFiles string
		if GetConfig().Debug {
			packageAssetFolderPath := path.Join(GetConfig().DstDir, packageAssetFolderName)
			MustCreateFolderIfNotExists(packageAssetFolderPath, os.ModePerm)
			// Create folder for assets generated during this run:
			thisRunAssetFolderName := ToLowerCamel(idl.Name) + "_" + ts.Format(FilenameTimeFormat)
			thisRunAssetFolderPath := path.Join(packageAssetFolderPath, thisRunAssetFolderName)
			// Create a new assets folder inside the main assets folder:
			MustCreateFolderIfNotExists(thisRunAssetFolderPath, os.ModePerm)
			dstDirForFiles = thisRunAssetFolderPath
		} else {
			if GetConfig().DstDir == generatedDir {
				dstDirForFiles = filepath.Join(GetConfig().DstDir, packageAssetFolderName)
			} else {
				dstDirForFiles = GetConfig().DstDir
			}
		}
		MustCreateFolderIfNotExists(dstDirForFiles, os.ModePerm)

		files, err := GenerateClientFromProgramIDL(idl)
		if err != nil {
			panic(err)
		}

		{
			mdf := &modfile.File{}
			mdf.AddModuleStmt(GetConfig().ModPath)

			mdf.AddNewRequire("github.com/gagliardetto/solana-go", "v1.5.0", false)
			mdf.AddNewRequire("github.com/gagliardetto/binary", "v0.7.1", false)
			mdf.AddNewRequire("github.com/gagliardetto/treeout", "v0.1.4", false)
			mdf.AddNewRequire("github.com/gagliardetto/gofuzz", "v1.2.2", false)
			mdf.AddNewRequire("github.com/stretchr/testify", "v1.6.1", false)
			mdf.AddNewRequire("github.com/davecgh/go-spew", "v1.1.1", false)
			mdf.Cleanup()

			callbacks = append(callbacks, func() {
				Ln()
				Ln(Bold("Don't forget to import the necessary dependencies!"))
				Ln()
				for _, v := range mdf.Require {
					Sfln(
						"	go get %s@%s",
						v.Mod.Path,
						v.Mod.Version,
					)
				}
				Ln()
			})

			if GetConfig().ModPath != "" {
				mfBytes, err := mdf.Format()
				if err != nil {
					panic(err)
				}
				gomodFilepath := filepath.Join(dstDirForFiles, "go.mod")
				Sfln(
					"[%s] %s",
					Lime(Checkmark),
					MustAbs(gomodFilepath),
				)
				// Write `go.mod` file:
				err = ioutil.WriteFile(gomodFilepath, mfBytes, 0666)
				if err != nil {
					panic(err)
				}
			}
		}

		for _, file := range files {
			// err := file.Render(os.Stdout)
			// if err != nil {
			// 	panic(err)
			// }

			file.File.HeaderComment("Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.")
			{
				// Save assets:
				assetFileName := file.Name + ".go"
				assetFilepath := path.Join(dstDirForFiles, assetFileName)

				// Create file:
				goFile, err := os.Create(assetFilepath)
				if err != nil {
					panic(err)
				}
				defer goFile.Close()

				// Write generated code file:
				Sfln(
					"[%s] %s",
					Lime(Checkmark),
					MustAbs(assetFilepath),
				)
				err = file.File.Render(goFile)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func FormatSighash(buf []byte) string {
	elems := make([]string, 0)
	for _, v := range buf {
		elems = append(elems, strconv.Itoa(int(v)))
	}

	return "[" + strings.Join(elems, ", ") + "]"
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
		// register complex enums:
		for _, typ := range idl.Types {
			registerComplexEnums(&idl, typ)
		}
		for _, typ := range idl.Accounts {
			registerComplexEnums(&idl, typ)
		}
	}

	{
		file := NewGoFile(idl.Name, true)
		// Declare types from IDL:
		for _, typ := range idl.Types {
			file.Add(genTypeDef(&idl, false, typ))
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
			// generate type definition:
			file.Add(genTypeDef(&idl, GetConfig().TypeID == TypeIDAnchor, acc))
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
		insAccountsExportedName := ToCamel(fmt.Sprintf("%s_accounts", instruction.Name))
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
					fieldsGroup.Add(genField(arg, true)).
						Add(func() Code {
							if arg.Type.IsIdlTypeOption() {
								return Tag(map[string]string{
									"bin": "optional",
								})
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
					"bin": "-",
				})
			})

			code.Line().Line()

			// Generate Instruction Account Struct
			code.Type().Id(insAccountsExportedName).StructFunc(func(fieldsGroup *Group) {
				for _, accountItem := range instruction.Accounts {
					fieldsGroup.Add(genAccountField(accountItem))
				}
			})

			file.Add(code.Line())
		}

		{
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
							pureVarName := getSysVarName(account.Name)
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
			// Declare method to get Account struct from AccountMetasSlice
			code := Empty().Line().Line()
			code.
				Func().
				Params(Id("inst").Op("*").Id(insExportedName)).
				Id("Get" + insAccountsExportedName).Params().Op("*").Id(insAccountsExportedName).
				BlockFunc(func(body *Group) {
					body.Id("res").Op(":=").Op("&").Id(insAccountsExportedName).Block()
					instruction.Accounts.Walk("", nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, account *IdlAccount) bool {
						// check if not nil before accessing "PublicKey"
						body.If(Id("inst").Dot("AccountMetaSlice").Index(Lit(index)).Op("!=").Nil()).BlockFunc(func(group *Group) {
							field := group.Id("res")
							if parentGroupPath != "" {
								field = field.Dot(ToCamel(parentGroupPath))
							}
							field.Dot(ToCamel(account.Name)).Op("=").Add(Id("inst").Dot("AccountMetaSlice").Index(Lit(index)).Dot("PublicKey"))
						})
						return true
					})
					body.Return().Op("res")
				})

			file.Add(code.Line().Line())
		}

		{
			// Declare methods that set the parameters of the instruction:
			code := Empty()
			for _, arg := range instruction.Args {
				exportedArgName := ToCamel(arg.Name)

				code.Line().Line()
				name := "Set" + exportedArgName
				code.Commentf("%s sets the %q parameter.", name, arg.Name).Line()
				for _, doc := range arg.Docs {
					code.Comment(doc).Line()
				}

				code.Func().Params(Id("inst").Op("*").Id(insExportedName)).Id(name).
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
							"bin": "-",
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
										Op("=").Id(ToLowerCamel(builderStructName)).Dot(formatAccountAccessorName("Get", exportedAccountName)).Call()

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

					var typeIDCode Code

					GetConfig().TypeID.
						On(
							TypeIDNameSlice{
								TypeIDUvarint32,
							},
							func() {
								typeIDCode = Qual(PkgDfuseBinary, "TypeIDFromUvarint32").Call(Id("Instruction_" + insExportedName))
							},
						).
						On(
							TypeIDNameSlice{
								TypeIDUint32,
							},
							func() {
								typeIDCode = Qual(PkgDfuseBinary, "TypeIDFromUint32").Call(Id("Instruction_"+insExportedName), Qual("encoding/binary", "LittleEndian"))
							},
						).
						On(
							TypeIDNameSlice{
								TypeIDUint8,
							},
							func() {
								typeIDCode = Qual(PkgDfuseBinary, "TypeIDFromUint8").Call(Id("Instruction_" + insExportedName))
							},
						).
						On(
							TypeIDNameSlice{
								TypeIDAnchor,
							},
							func() {
								typeIDCode = Id("Instruction_" + insExportedName)
							},
						).
						On(
							TypeIDNameSlice{
								TypeIDNoType,
							},
							func() {
								// TODO
							},
						)

					body.Return().Op("&").Id("Instruction").Values(
						Dict{
							Id("BaseVariant"): Qual(PkgDfuseBinary, "BaseVariant").Values(
								Dict{
									Id("TypeID"): typeIDCode,
									Id("Impl"):   Id("inst"),
								},
							),
						},
					)
				})
			file.Add(code.Line())
		}

		{
			// Declare `ValidateAndBuild` method on instruction:
			code := Empty()

			code.Line().Line().
				Comment("ValidateAndBuild validates the instruction parameters and accounts;").
				Line().
				Comment("if there is a validation error, it returns the error.").
				Line().
				Comment("Otherwise, it builds and returns the instruction.").
				Line().
				Func().Params(Id("inst").Id(insExportedName)).Id("ValidateAndBuild").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
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
					body.If(
						Err().Op(":=").Id("inst").Dot("Validate").Call(),
						Err().Op("!=").Nil(),
					).Block(
						Return(Nil(), Err()),
					)

					body.Return(Id("inst").Dot("Build").Call(), Nil())
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

					body.Comment("Check whether all (required) accounts are set:")
					body.BlockFunc(func(accountValidationBlock *Group) {
						instruction.Accounts.Walk("", nil, nil, func(groupPath string, accountIndex int, parentGroup *IdlAccounts, ia *IdlAccount) bool {
							exportedAccountName := ToCamel(filepath.Join(groupPath, ia.Name))

							if ia.Optional {
								accountValidationBlock.Line().Commentf(
									"[%v] = %s is optional",
									accountIndex,
									exportedAccountName,
								).Line()
							} else {
								accountValidationBlock.If(Id("inst").Dot("AccountMetaSlice").Index(Lit(accountIndex)).Op("==").Nil()).Block(
									Return(Qual("errors", "New").Call(Lit(Sf("accounts.%s is not set", exportedAccountName)))),
								)
							}

							return true
						})
					})

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

									instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit(Sf("Params[len=%v]", len(instruction.Args)))).Dot("ParentFunc").Parens(Func().Parens(Id("paramsBranch").Qual(PkgTreeout, "Branches")).BlockFunc(func(paramsBranchGroup *Group) {
										longest := treeFindLongestNameFromFields(instruction.Args)
										for _, arg := range instruction.Args {
											exportedArgName := ToCamel(arg.Name)
											paramsBranchGroup.Id("paramsBranch").Dot("Child").
												Call(
													Qual(PkgFormat, "Param").Call(
														Lit(strings.Repeat(" ", longest-len(exportedArgName))+exportedArgName+StringIf(arg.Type.IsIdlTypeOption(), " (OPT)")),
														Add(CodeIf(!arg.Type.IsIdlTypeOption() && !isComplexEnum(arg.Type), Op("*"))).Id("inst").Dot(exportedArgName),
													),
												)
										}
									}))

									instructionBranchGroup.Line().Comment("Accounts of the instruction:")

									instructionBranchGroup.Id("instructionBranch").Dot("Child").Call(Lit(Sf("Accounts[len=%v]", instruction.Accounts.NumAccounts()))).Dot("ParentFunc").Parens(
										Func().Parens(Id("accountsBranch").Qual(PkgTreeout, "Branches")).BlockFunc(func(accountsBranchGroup *Group) {

											longest := treeFindLongestNameFromAccounts(instruction.Accounts)
											instruction.Accounts.Walk("", nil, nil, func(groupPath string, accountIndex int, parentGroup *IdlAccounts, ia *IdlAccount) bool {

												cleanedName := treeFormatAccountName(ia.Name)

												exportedAccountName := filepath.Join(groupPath, cleanedName)

												access := Id("accountsBranch").Dot("Child").Call(Qual(PkgFormat, "Meta").Call(Lit(strings.Repeat(" ", longest-len(exportedAccountName))+exportedAccountName), Id("inst").Dot("AccountMetaSlice").Dot("Get").Call(Lit(accountIndex))))
												accountsBranchGroup.Add(access)
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
			file.Add(
				genMarshalWithEncoder_struct(
					&idl,
					false,
					insExportedName,
					"",
					instruction.Args,
					true,
				),
			)
		}

		{
			// Declare `UnmarshalWithDecoder(decoder *bin.Decoder) error` method on instruction:
			file.Add(
				genUnmarshalWithDecoder_struct(
					&idl,
					false,
					insExportedName,
					"",
					instruction.Args,
					bin.TypeID{},
				))
		}

		{
			// Declare instruction initializer func:
			paramNames := []string{}
			for _, arg := range instruction.Args {
				paramNames = append(paramNames, arg.Name)
			}
			code := Empty()
			name := "New" + insExportedName + "Instruction"
			code.Commentf("%s declares a new %s instruction with the provided parameters and accounts.", name, insExportedName)
			code.Line()
			code.Func().Id(name).
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						{
							for argIndex, arg := range instruction.Args {
								paramNames = append(paramNames, arg.Name)
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

								if SliceContains(paramNames, accountName) {
									accountName = accountName + "Account"
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
						results.Op("*").Id(insExportedName)
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
								// TODO
							}

							if SliceContains(paramNames, accountName) {
								accountName = accountName + "Account"
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
														}()).Id(formatAccountAccessorName("Set", exportedAccountName)).Call(Id(accountName))

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
								builder.Op(".").Line().Id(formatAccountAccessorName("Set", ToCamel(account.Name))).Call(Id(accountName))
							}

							return true
						})
					}

					// builder.Op(".").Line().Id("Build").Call()
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
	code := Empty()

	{
		code.Line().Line()
		name := formatAccountAccessorName("Set", exportedAccountName)
		code.Commentf("%s sets the %q account.", name, account.Name).Line()
		for _, doc := range account.Docs {
			code.Comment(doc).Line()
		}
		// Create account setters:
		code.Func().Params(Id("inst").Op("*").Id(receiverTypeName)).Id(name).
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
	}

	{ // Create account getters:
		code.Line().Line()
		name := formatAccountAccessorName("Get", exportedAccountName)
		if account.Optional {
			code.Commentf("%s gets the %q account (optional).", name, account.Name).Line()
		} else {
			code.Commentf("%s gets the %q account.", name, account.Name).Line()
		}
		for _, doc := range account.Docs {
			code.Comment(doc).Line()
		}
		code.Func().Params(Id("inst").Op("*").Id(receiverTypeName)).Id(name).
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
				body.Return(Id("inst").Dot("AccountMetaSlice").Dot("Get").Call(Lit(index)))
			})
	}

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
			// TODO: check for pointer instead.
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
		GetConfig().TypeID.
			On(
				TypeIDNameSlice{
					TypeIDUvarint32,
					TypeIDUint32,
					TypeIDUint8,
				},
				func() {

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
									switch GetConfig().TypeID {
									case TypeIDUvarint32, TypeIDUint32:
										ins.Uint32().Op("=").Iota().Line()
									case TypeIDUint8:
										ins.Uint8().Op("=").Iota().Line()
									}
								}
								gr.Add(ins.Line().Line())
							}
						}),
					)
					file.Add(code.Line())

				},
			).
			On(
				TypeIDNameSlice{
					TypeIDAnchor,
				},
				func() {
					code := Empty()
					code.Var().Parens(
						DoGroup(func(gr *Group) {
							for _, instruction := range idl.Instructions {
								insExportedName := ToCamel(instruction.Name)

								ins := Empty().Line()
								for _, doc := range instruction.Docs {
									ins.Comment(doc).Line()
								}
								toBeHashed := sighash.ToSnakeForSighash(instruction.Name)
								if GetConfig().Debug {
									ins.Comment(Sf(`hash("%s:%s")`, bin.SIGHASH_GLOBAL_NAMESPACE, toBeHashed)).Line()
								}
								ins.Id("Instruction_" + insExportedName)

								ins.Op("=").Qual(PkgDfuseBinary, "TypeID").Call(
									Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
										sighash := bin.SighashTypeID(bin.SIGHASH_GLOBAL_NAMESPACE, toBeHashed)
										for _, byteVal := range sighash[:] {
											byteGroup.Lit(int(byteVal))
										}
									}).Op("}"),
								)
								gr.Add(ins.Line().Line())
							}
						}),
					)
					file.Add(code.Line())
				},
			).
			On(
				TypeIDNameSlice{
					TypeIDNoType,
				},
				func() {
					// TODO
				},
			)
	}
	{
		// Declare `InstructionIDToName` function:
		GetConfig().TypeID.
			On(
				TypeIDNameSlice{
					TypeIDUvarint32,
					TypeIDUint32,
					TypeIDUint8,
				},
				func() {
					code := Empty()
					code.Comment("InstructionIDToName returns the name of the instruction given its ID.").Line()
					code.Func().Id("InstructionIDToName").
						Params(
							func() Code {
								switch GetConfig().TypeID {
								case TypeIDUvarint32, TypeIDUint32:
									return Id("id").Uint32()
								case TypeIDUint8:
									return Id("id").Uint8()
								}
								return nil
							}(),
						).
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
				},
			).
			On(
				TypeIDNameSlice{
					TypeIDAnchor,
				},
				func() {
					code := Empty()
					code.Comment("InstructionIDToName returns the name of the instruction given its ID.").Line()
					code.Func().Id("InstructionIDToName").
						Params(Id("id").Qual(PkgDfuseBinary, "TypeID")).
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
				},
			).
			On(
				TypeIDNameSlice{
					TypeIDNoType,
				},
				func() {
					// TODO
				},
			)
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
			GetConfig().TypeID.
				On(
					TypeIDNameSlice{
						TypeIDUvarint32,
						TypeIDUint32,
						TypeIDUint8,
					},
					func() {

						code := Empty()
						code.Var().Id("InstructionImplDef").Op("=").Qual(PkgDfuseBinary, "NewVariantDefinition").
							Parens(DoGroup(func(call *Group) {
								call.Line()

								switch GetConfig().TypeID {
								case TypeIDUvarint32:
									call.Qual(PkgDfuseBinary, "Uvarint32TypeIDEncoding").Op(",").Line()
								case TypeIDUint32:
									call.Qual(PkgDfuseBinary, "Uint32TypeIDEncoding").Op(",").Line()
								case TypeIDUint8:
									call.Qual(PkgDfuseBinary, "Uint8TypeIDEncoding").Op(",").Line()
								}

								call.Index().Qual(PkgDfuseBinary, "VariantType").
									BlockFunc(func(variantBlock *Group) {
										for _, instruction := range idl.Instructions {
											// NOTE: using `ToCamel` here:
											insName := ToCamel(instruction.Name)
											insExportedName := ToCamel(instruction.Name)
											variantBlock.Block(
												List(Lit(insName), Parens(Op("*").Id(insExportedName)).Parens(Nil())).Op(","),
											).Op(",")
										}
									}).Op(",").Line()
							}))
						file.Add(code.Line())

					},
				).
				On(
					TypeIDNameSlice{
						TypeIDAnchor,
					},
					func() {
						code := Empty()
						code.Var().Id("InstructionImplDef").Op("=").Qual(PkgDfuseBinary, "NewVariantDefinition").
							Parens(DoGroup(func(call *Group) {
								call.Line()
								call.Qual(PkgDfuseBinary, "AnchorTypeIDEncoding").Op(",").Line()

								call.Index().Qual(PkgDfuseBinary, "VariantType").
									BlockFunc(func(variantBlock *Group) {
										for _, instruction := range idl.Instructions {
											// NOTE: using `ToSnakeForSighash` here (necessary for sighash computing from instruction name)
											insName := sighash.ToSnakeForSighash(instruction.Name)
											insExportedName := ToCamel(instruction.Name)
											variantBlock.Block(
												List(Lit(insName), Parens(Op("*").Id(insExportedName)).Parens(Nil())).Op(","),
											).Op(",")
										}
									}).Op(",").Line()
							}))
						file.Add(code.Line())
					},
				).
				On(
					TypeIDNameSlice{
						TypeIDNoType,
					},
					func() {
						// TODO
					},
				)

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

					GetConfig().TypeID.
						On(
							TypeIDNameSlice{
								TypeIDUvarint32,
								TypeIDUint32,
								TypeIDUint8,
							},
							func() {

								switch GetConfig().TypeID {
								case TypeIDUvarint32:
									body.Err().Op(":=").Id("encoder").Dot("WriteUVarInt").Call(Id("inst").Dot("TypeID").Dot("Uvarint32").Call())
								case TypeIDUint32:
									body.Err().Op(":=").Id("encoder").Dot("WriteUint32").Call(Id("inst").Dot("TypeID").Dot("Uint32").Call(), Qual("encoding/binary", "LittleEndian"))
								case TypeIDUint8:
									body.Err().Op(":=").Id("encoder").Dot("WriteUint8").Call(Id("inst").Dot("TypeID").Dot("Uint8").Call())
								}

							},
						).
						On(
							TypeIDNameSlice{
								TypeIDAnchor,
							},
							func() {
								body.Err().Op(":=").Id("encoder").Dot("WriteBytes").Call(Id("inst").Dot("TypeID").Dot("Bytes").Call(), False())
							},
						).
						On(
							TypeIDNameSlice{
								TypeIDNoType,
							},
							func() {
								// TODO
							},
						)

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

// formatAccountAccessorName formats a name for a function that
// either gets or sets an account.
// If the RemoveAccountSuffix config flag is set, and the name already
// has an "Account" suffix, then another "Account" suffix is NOT added.
// E.g. ("Set", "Foo") => "SetFooAccount"
// E.g. ("Set", "BarAccount") => "SetBarAccount"
func formatAccountAccessorName(prefix, name string) string {
	endsWithAccount := strings.HasSuffix(strings.ToLower(name), "account")
	if !conf.RemoveAccountSuffix || !endsWithAccount {
		return prefix + name + "Account"
	}
	return prefix + name
}

func treeFindLongestNameFromFields(fields []IdlField) (ln int) {
	for _, v := range fields {
		if len(v.Name) > ln {
			ln = len(v.Name)
		}
	}
	return
}

func treeFindLongestNameFromAccounts(accounts IdlAccountItemSlice) (ln int) {
	accounts.Walk("", nil, nil, func(groupPath string, accountIndex int, parentGroup *IdlAccounts, ia *IdlAccount) bool {

		cleanedName := treeFormatAccountName(ia.Name)

		exportedAccountName := filepath.Join(groupPath, cleanedName)
		if len(exportedAccountName) > ln {
			ln = len(exportedAccountName)
		}

		return true
	})
	return
}

func treeFormatAccountName(name string) string {
	cleanedName := name
	if isSysVar(name) {
		cleanedName = strings.TrimSuffix(getSysVarName(name), "Pubkey")
	}
	if len(cleanedName) > len("account") {
		if strings.HasSuffix(cleanedName, "account") {
			cleanedName = strings.TrimSuffix(cleanedName, "account")
		} else if strings.HasSuffix(cleanedName, "Account") {
			cleanedName = strings.TrimSuffix(cleanedName, "Account")
		}
	}
	return cleanedName
}
