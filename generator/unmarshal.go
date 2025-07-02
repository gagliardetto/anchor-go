package generator

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

func formatComplexEnumVariantTypeName(enumTypeName string, variantName string) string {
	return fmt.Sprintf("%s_%s", tools.ToCamelUpper(enumTypeName), tools.ToCamelUpper(variantName))
}

func formatSimpleEnumVariantName(variantName string, enumTypeName string) string {
	return fmt.Sprintf("%s_%s", tools.ToCamelUpper(enumTypeName), tools.ToCamelUpper(variantName))
}

func FormatTupleItemName(index int) string {
	return tools.ToCamelUpper(fmt.Sprintf("V%d", index))
}

func formatEnumContainerName(enumTypeName string) string {
	return tools.ToCamelLower(enumTypeName) + "EnumContainer"
}

func formatInterfaceMethodName(enumTypeName string) string {
	return "is" + tools.ToCamelUpper(enumTypeName)
}

func formatDiscriminatorName(kind string, exportedAccountName string) string {
	// trim prefix or suffix "Account" or "Event" from exportedAccountName
	exportedAccountName = tools.ToCamelUpper(exportedAccountName)

	// // TODO: sometimes there's accounts/events like this:
	// // - "Fund"
	// // - "FundAccount"
	// // This will create a name collision and fail to compile because
	// // we remove the "Account" or "Event" suffix from the second one,
	// // so there's a duplicate name "Fund".
	// exportedAccountName = strings.TrimSuffix(exportedAccountName, "Account")
	// exportedAccountName = strings.TrimSuffix(exportedAccountName, "Event")
	// exportedAccountName = strings.TrimPrefix(exportedAccountName, "Account")
	// exportedAccountName = strings.TrimPrefix(exportedAccountName, "Event")

	return kind + "_" + tools.ToCamelUpper(exportedAccountName)
}

func FormatAccountDiscriminatorName(exportedAccountName string) string {
	return formatDiscriminatorName("Account", exportedAccountName)
}

func FormatEventDiscriminatorName(exportedEventName string) string {
	return formatDiscriminatorName("Event", exportedEventName)
}

func FormatInstructionDiscriminatorName(exportedInstructionName string) string {
	return formatDiscriminatorName("Instruction", exportedInstructionName)
}

func formatBuilderFuncName(insExportedName string) string {
	return "New" + insExportedName + "InstructionBuilder"
}

func formatEnumParserName(enumTypeName string) string {
	return "Decode" + enumTypeName
}

func formatEnumEncoderName(enumTypeName string) string {
	return "Encode" + enumTypeName
}

func gen_UnmarshalWithDecoder_struct(
	idl_ *idl.Idl,
	withDiscriminator bool,
	receiverTypeName string,
	discriminatorName string,
	fields idl.IdlDefinedFields,
) Code {
	code := Empty()
	{
		// Declare UnmarshalWithDecoder
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id("UnmarshalWithDecoder").
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("decoder").Op("*").Qual(PkgBinary, "Decoder")
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
						discReadBody.List(Id("discriminator"), Err()).Op(":=").Id("decoder").Dot("ReadDiscriminator").Call()
						discReadBody.If(Err().Op("!=").Nil()).Block(
							Return(Err()),
						)
						discReadBody.If(Op("!").Id("discriminator").Dot("Equal").Call(Id(discriminatorName).Index(Op(":")))).Block(
							Return(
								Qual("fmt", "Errorf").Call(
									Line().Lit("wrong discriminator: wanted %s, got %s"),
									Line().Id(discriminatorName).Index(Op(":")),
									Line().Qual("fmt", "Sprint").Call(Id("discriminator").Index(Op(":"))),
								),
							),
						)
					})
				}

				switch fields := fields.(type) {
				case idl.IdlDefinedFieldsNamed:
					gen_unmarshal_DefinedFieldsNamed(body, fields)
				case idl.IdlDefinedFieldsTuple:
					convertedFields := tupleToFieldsNamed(fields)
					gen_unmarshal_DefinedFieldsNamed(body, convertedFields)
				case nil:
					// No fields, just an empty struct.
					// TODO: should we panic here?
				default:
					panic(fmt.Sprintf("unexpected fields type: %T", fields))
				}

				body.Return(Nil())
			})
	}
	{
		code.Line().Line()
		// func (obj *<type>) Unmarshal(buf []byte) (err error) {
		// 	return obj.UnmarshalWithDecoder(bin.NewBorshDecoder(buf))
		// }
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id("Unmarshal").
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("buf").Index().Byte()
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
				body.Err().Op(":=").Id("obj").Dot("UnmarshalWithDecoder").Call(
					Qual(PkgBinary, "NewBorshDecoder").Call(Id("buf")),
				)
				body.If(Err().Op("!=").Nil()).Block(
					// If there was an error, return it.
					Return(
						Qual("fmt", "Errorf").Call(
							Lit("error while unmarshaling "+receiverTypeName+": %w"),
							Err(),
						),
					),
				)
				body.Return(
					Nil(), // No error.
				)
			})
	}
	{
		code.Line().Line()
		// func Unmarshal<type>(buf []byte) (<type>, error) {
		// 	obj := new(<type>)
		// 	err := obj.Unmarshal(buf)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	return obj, nil
		// }
		code.Func().Id("Unmarshal" + receiverTypeName).
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("buf").Index().Byte()
				}),
			).
			Params(
				ListFunc(func(results *Group) {
					// Results:
					results.Op("*").Id(receiverTypeName)
					results.Error()
				}),
			).
			BlockFunc(func(body *Group) {
				// Body:
				body.Id("obj").Op(":=").New(Id(receiverTypeName))
				body.Err().Op(":=").Id("obj").Dot("Unmarshal").Call(Id("buf"))
				body.If(Err().Op("!=").Nil()).Block(
					Return(
						Nil(),
						Err(),
					),
				)
				body.Return(
					Id("obj"),
					Nil(), // No error.
				)
			})
	}
	return code
}

func tupleToFieldsNamed(
	tuple idl.IdlDefinedFieldsTuple,
) idl.IdlDefinedFieldsNamed {
	fields := make(idl.IdlDefinedFieldsNamed, len(tuple))
	for i, item := range tuple {
		tupleItemName := FormatTupleItemName(i)
		fields[i] = idl.IdlField{
			Name: tupleItemName,
			Ty:   item,
		}
	}
	return fields
}

func gen_unmarshal_DefinedFieldsNamed(
	body *Group,
	fields idl.IdlDefinedFieldsNamed,
) {
	for _, field := range fields {
		exportedArgName := tools.ToCamelUpper(field.Name)
		if IsOption(field.Ty) || IsCOption(field.Ty) {
			body.Commentf("Deserialize `%s` (optional):", exportedArgName)
		} else {
			body.Commentf("Deserialize `%s`:", exportedArgName)
		}

		if isComplexEnum(field.Ty) || (IsArray(field.Ty) && isComplexEnum(field.Ty.(*idltype.Array).Type)) || (IsVec(field.Ty) && isComplexEnum(field.Ty.(*idltype.Vec).Vec)) {
			// TODO: this assumes this cannot be an option;
			// - check whether this is an option?
			switch field.Ty.(type) {
			case *idltype.Defined:
				enumName := field.Ty.(*idltype.Defined).Name
				body.BlockFunc(func(argBody *Group) {
					{
						argBody.Var().Err().Error()
						argBody.List(
							Id("obj").Dot(exportedArgName),
							Err(),
						).Op("=").Id(formatEnumParserName(enumName)).Call(Id("decoder"))
					}
					argBody.If(
						Err().Op("!=").Nil(),
					).Block(
						Return(Err()),
					)
				})
			case *idltype.Array:
				enumTypeName := field.Ty.(*idltype.Array).Type.(*idltype.Defined).Name
				body.BlockFunc(func(argBody *Group) {
					// Read the array items:
					argBody.For(
						Id("i").Op(":=").Lit(0),
						Id("i").Op("<").Len(Id("obj").Dot(exportedArgName)),
						Id("i").Op("++"),
					).BlockFunc(func(forBody *Group) {
						forBody.List(
							Id("obj").Dot(exportedArgName).Index(Id("i")),
							Err(),
						).Op("=").Id(formatEnumParserName(enumTypeName)).Call(Id("decoder"))
						forBody.If(Err().Op("!=").Nil()).Block(
							Return(
								Qual(PkgAnchorGoErrors, "NewField").Call(
									Lit(exportedArgName),
									Qual(PkgAnchorGoErrors, "NewIndex").Call(
										Id("i"),
										Err(),
									),
								),
							),
						)
					})
				})
			case *idltype.Vec:
				enumTypeName := field.Ty.(*idltype.Vec).Vec.(*idltype.Defined).Name
				body.BlockFunc(func(argBody *Group) {
					// Read the vector length:
					argBody.List(Id("vecLen"), Err()).Op(":=").Id("decoder").Dot("ReadLength").Call()
					argBody.If(Err().Op("!=").Nil()).Block(
						Return(
							Qual(PkgAnchorGoErrors, "NewField").Call(
								Lit(exportedArgName),
								Qual("fmt", "Errorf").Call(
									Lit("error while reading vector length: %w"),
									Err(),
								),
							),
						),
					)
					// Create the vector:
					argBody.Id("obj").Dot(exportedArgName).Op("=").Make(Index().Id(enumTypeName), Id("vecLen"))
					// Read the vector items:
					argBody.For(
						Id("i").Op(":=").Lit(0),
						Id("i").Op("<").Id("vecLen"),
						Id("i").Op("++"),
					).BlockFunc(func(forBody *Group) {
						forBody.List(
							Id("obj").Dot(exportedArgName).Index(Id("i")),
							Err(),
						).Op("=").Id(formatEnumParserName(enumTypeName)).Call(Id("decoder"))
						forBody.If(Err().Op("!=").Nil()).Block(
							Return(
								Qual(PkgAnchorGoErrors, "NewField").Call(
									Lit(exportedArgName),
									Qual(PkgAnchorGoErrors, "NewIndex").Call(
										Id("i"),
										Err(),
									),
								),
							),
						)
					})
				})
			}
		} else {
			if IsOption(field.Ty) || IsCOption(field.Ty) {
				var optionalityReaderName string
				switch {
				case IsOption(field.Ty):
					optionalityReaderName = "ReadOption"
				case IsCOption(field.Ty):
					optionalityReaderName = "ReadCOption"
				}

				body.BlockFunc(func(optGroup *Group) {
					// if nil:
					optGroup.List(Id("ok"), Err()).Op(":=").Id("decoder").Dot(optionalityReaderName).Call()
					optGroup.If(Err().Op("!=").Nil()).Block(
						Return(
							Qual(PkgAnchorGoErrors, "NewOption").Call(
								Lit(exportedArgName),
								Qual("fmt", "Errorf").Call(
									Lit("error while reading optionality: %w"),
									Err(),
								),
							),
						),
					)
					optGroup.If(Id("ok")).Block(
						Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName)),
						If(Err().Op("!=").Nil()).Block(
							Return(
								Qual(PkgAnchorGoErrors, "NewField").Call(
									Lit(exportedArgName),
									Err(),
								),
							),
						),
					)
				})
			} else {
				body.Err().Op("=").Id("decoder").Dot("Decode").Call(Op("&").Id("obj").Dot(exportedArgName))
				body.If(Err().Op("!=").Nil()).Block(
					Return(
						Qual(PkgAnchorGoErrors, "NewField").Call(
							Lit(exportedArgName),
							Err(),
						),
					),
				)
			}
		}
	}
}
