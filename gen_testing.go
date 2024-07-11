package main

import (
	. "github.com/dave/jennifer/jen"
	. "github.com/gagliardetto/utilz"
	"strings"
)

func isAnyFieldComplexEnum(envelopes ...IdlField) bool {
	for _, v := range envelopes {
		if isComplexEnum(v.Type) {
			return true
		}
	}
	return false
}

func countFieldComplexEnum(envelopes ...IdlField) int {
	var count int
	for _, v := range envelopes {
		if isComplexEnum(v.Type) {
			count++
		}
	}
	return count
}

func genTestingFuncs(idl IDL) ([]*FileWrapper, error) {

	files := make([]*FileWrapper, 0)
	{
		file := NewGoFile(idl.Metadata.Name, true)
		// Declare testing tools:
		{
			code := Empty()
			code.Func().Id("encodeT").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("data").Interface()
						params.Id("buf").Op("*").Qual("bytes", "Buffer")
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
					body.If(
						Err().Op(":=").Qual(PkgDfuseBinary, GetConfig().Encoding._NewEncoder()).Call(Id("buf")).Dot("Encode").Call(Id("data")),
						Err().Op("!=").Nil(),
					).Block(
						Return(Qual("fmt", "Errorf").Call(Lit("unable to encode instruction: %w"), Err())),
					)
					body.Return(Nil())
				})
			file.Add(code.Line())
		}
		{
			code := Empty()
			code.Func().Id("decodeT").
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("dst").Interface()
						params.Id("data").Index().Byte()
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
					body.Return(Qual(PkgDfuseBinary, GetConfig().Encoding._NewDecoder()).Call(Id("data")).Dot("Decode").Call(Id("dst")))
				})
			file.Add(code.Line())
		}
		////
		files = append(files, &FileWrapper{
			Name: "testing_utils",
			File: file,
		})
	}
	for _, instruction := range idl.Instructions {
		file := NewGoFile(idl.Metadata.Name, true)
		insExportedName := ToCamel(instruction.Name)
		{
			// Declare test: encode, decode:
			code := Empty()
			code.Func().Id("TestEncodeDecode_" + insExportedName).
				Params(
					ListFunc(func(params *Group) {
						// Parameters:
						params.Id("t").Op("*").Qual("testing", "T")
					}),
				).
				Params(
					ListFunc(func(results *Group) {
						// Results:
					}),
				).
				BlockFunc(func(body *Group) {
					// Body:
					body.Id("fu").Op(":=").Qual(PkgGoFuzz, "New").Call().Dot("NilChance").Call(Lit(0))

					body.For(
						Id("i").Op(":=").Lit(0),
						Id("i").Op("<").Lit(1),
						Id("i").Op("++"),
					).BlockFunc(func(forGroup *Group) {
						forGroup.Id("t").Dot("Run").Call(
							Lit(insExportedName).Op("+").Qual("strconv", "Itoa").Call(Id("i")),
							DoGroup(func(fnGroup *Group) {
								fnGroup.Func().Params(Id("t").Op("*").Qual("testing", "T")).Block(
									BlockFunc(func(tFunGroup *Group) {

										if isAnyFieldComplexEnum(instruction.Args...) {
											genTestWithComplexEnum(tFunGroup, insExportedName, instruction, idl)
											// TODO: populate complex enum
										} else {
											genTestNOComplexEnum(tFunGroup, insExportedName, instruction)
										}

									}),
								)
							}),
						)
					})

				})

			file.Add(code.Line())
		}
		////
		files = append(files, &FileWrapper{
			Name: strings.ToLower(insExportedName) + "_test",
			File: file,
		})
	}

	return files, nil
}

func genTestNOComplexEnum(tFunGroup *Group, insExportedName string, instruction IdlInstruction) {
	tFunGroup.Id("params").Op(":=").New(Id(insExportedName))

	tFunGroup.Id("fu").Dot("Fuzz").Call(Id("params"))
	tFunGroup.Id("params").Dot("AccountMetaSlice").Op("=").Nil()

	tFunGroup.Id("buf").Op(":=").New(Qual("bytes", "Buffer"))
	tFunGroup.Id("err").Op(":=").Id("encodeT").Call(Op("*").Id("params"), Id("buf"))
	tFunGroup.Qual(PkgTestifyRequire, "NoError").Call(Id("t"), Err())

	tFunGroup.Comment("//")

	tFunGroup.Id("got").Op(":=").New(Id(insExportedName))
	tFunGroup.Id("err").Op("=").Id("decodeT").Call(Id("got"), Id("buf").Dot("Bytes").Call())
	tFunGroup.Id("got").Dot("AccountMetaSlice").Op("=").Nil()
	tFunGroup.Qual(PkgTestifyRequire, "NoError").Call(Id("t"), Err())
	tFunGroup.Qual(PkgTestifyRequire, "Equal").Call(Id("t"), Id("params"), Id("got"))
}

func genTestWithComplexEnum(tFunGroup *Group, insExportedName string, instruction IdlInstruction, idl IDL) {
	// Create a test for each complex enum argument:
	for _, arg := range instruction.Args {
		if !isComplexEnum(arg.Type) {
			continue
		}
		exportedArgName := ToCamel(arg.Name)

		tFunGroup.BlockFunc(func(enumBlock *Group) {

			enumName := arg.Type.GetIdlTypeDefined().Defined.Name
			interfaceType := idl.Types.GetByName(enumName)
			for _, variant := range *interfaceType.Type.Variants {

				enumBlock.BlockFunc(func(variantBlock *Group) {
					variantBlock.Id("params").Op(":=").New(Id(insExportedName))

					variantBlock.Id("fu").Dot("Fuzz").Call(Id("params"))
					variantBlock.Id("params").Dot("AccountMetaSlice").Op("=").Nil()

					variantBlock.Id("tmp").Op(":=").New(Id(ToCamel(variant.Name)))
					variantBlock.Id("fu").Dot("Fuzz").Call(Id("tmp"))
					variantBlock.Id("params").Dot("Set" + exportedArgName).Call(Id("tmp"))

					variantBlock.Id("buf").Op(":=").New(Qual("bytes", "Buffer"))
					variantBlock.Id("err").Op(":=").Id("encodeT").Call(Op("*").Id("params"), Id("buf"))
					variantBlock.Qual(PkgTestifyRequire, "NoError").Call(Id("t"), Err())

					variantBlock.Comment("//")

					variantBlock.Id("got").Op(":=").New(Id(insExportedName))
					variantBlock.Id("err").Op("=").Id("decodeT").Call(Id("got"), Id("buf").Dot("Bytes").Call())
					variantBlock.Id("got").Dot("AccountMetaSlice").Op("=").Nil()
					variantBlock.Qual(PkgTestifyRequire, "NoError").Call(Id("t"), Err())
					variantBlock.Qual(PkgTestifyRequire, "Equal").Call(Id("t"), Id("params"), Id("got"))
				})
			}

		})
	}
}
