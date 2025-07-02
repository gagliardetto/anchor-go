package generator

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

func gen_MarshalWithEncoder_struct(
	idl_ *idl.Idl,
	withDiscriminator bool,
	receiverTypeName string,
	discriminatorName string,
	fields idl.IdlDefinedFields,
	checkNil bool,
) Code {
	code := Empty()
	{
		// Declare MarshalWithEncoder
		code.Func().Params(Id("obj").Id(receiverTypeName)).Id("MarshalWithEncoder").
			Params(
				ListFunc(func(params *Group) {
					// Parameters:
					params.Id("encoder").Op("*").Qual(PkgBinary, "Encoder")
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
				switch fields := fields.(type) {
				case idl.IdlDefinedFieldsNamed:
					gen_marshal_DefinedFieldsNamed(
						body,
						fields,
						checkNil,
						func(field idl.IdlField) *Statement {
							return Id("obj").Dot(tools.ToCamelUpper(field.Name))
						},
						"encoder",
						false, // returnNilErr
						func(field idl.IdlField) string {
							return tools.ToCamelUpper(field.Name)
						},
					)
				case idl.IdlDefinedFieldsTuple:
					convertedFields := tupleToFieldsNamed(fields)
					gen_marshal_DefinedFieldsNamed(
						body,
						convertedFields,
						checkNil,
						func(field idl.IdlField) *Statement {
							return Id("obj").Dot(tools.ToCamelUpper(field.Name))
						},
						"encoder",
						false, // returnNilErr
						func(field idl.IdlField) string {
							return tools.ToCamelUpper(field.Name)
						},
					)
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
		// also generate a
		// func (obj <type>) Marshal() ([]byte, error) {
		// 	return obj.MarshalWithEncoder(bin.NewBorshEncoder(buf))
		// }
		// func (obj <type>) Marshal() ([]byte, error) {
		// 	buf := new(bytes.Buffer)
		// enc := bin.NewBorshEncoder(buf)
		// err := enc.Encode(meta)
		// if err != nil {
		//   return nil, err
		// }
		// return buf.Bytes(), nil
		// }
		code.Func().Params(Id("obj").Id(receiverTypeName)).Id("Marshal").
			Params(
				ListFunc(func(results *Group) {
					// no parameters
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
				body.Id("buf").Op(":=").Qual("bytes", "NewBuffer").Call(Nil())
				body.Id("encoder").Op(":=").Qual(PkgBinary, "NewBorshEncoder").Call(Id("buf"))
				body.Err().Op(":=").Id("obj").Dot("MarshalWithEncoder").Call(Id("encoder"))
				body.If(Err().Op("!=").Nil()).Block(
					Return(
						Nil(),
						Qual("fmt", "Errorf").Call(
							Lit("error while encoding "+receiverTypeName+": %w"),
							Err(),
						),
					),
				)
				body.Return(
					Id("buf").Dot("Bytes").Call(),
					Nil(),
				)
			})
	}

	return code
}

func gen_marshal_DefinedFieldsNamed(
	body *Group,
	fields idl.IdlDefinedFieldsNamed,
	checkNil bool,
	nameFormatter func(field idl.IdlField) *Statement,
	encoderVariableName string,
	returnNilErr bool,
	traceNameFormatter func(field idl.IdlField) string,
) {
	for _, field := range fields {
		exportedArgName := traceNameFormatter(field)
		if IsOption(field.Ty) || IsCOption(field.Ty) {
			body.Commentf("Serialize `%s` (optional):", exportedArgName)
		} else {
			body.Commentf("Serialize `%s`:", exportedArgName)
		}

		if isComplexEnum(field.Ty) || (IsArray(field.Ty) && isComplexEnum(field.Ty.(*idltype.Array).Type)) || (IsVec(field.Ty) && isComplexEnum(field.Ty.(*idltype.Vec).Vec)) {
			switch field.Ty.(type) {
			case *idltype.Defined:
				enumTypeName := field.Ty.(*idltype.Defined).Name
				body.BlockFunc(func(argBody *Group) {
					argBody.Err().Op(":=").Id(formatEnumEncoderName(enumTypeName)).Call(Id(encoderVariableName), nameFormatter(field))
					argBody.If(
						Err().Op("!=").Nil(),
					).Block(
						ReturnFunc(
							func(returnBody *Group) {
								if returnNilErr {
									returnBody.Nil()
								}
								returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
									Lit(exportedArgName),
									Err(),
								)
							},
						),
					)
				})
			case *idltype.Array:
				enumTypeName := field.Ty.(*idltype.Array).Type.(*idltype.Defined).Name
				// TODO: handle array length, which is defined in the type.
				body.BlockFunc(func(argBody *Group) {
					argBody.For(
						Id("i").Op(":=").Lit(0),
						Id("i").Op("<").Len(nameFormatter(field)),
						Id("i").Op("++"),
					).BlockFunc(func(forBody *Group) {
						forBody.Err().Op(":=").Id(formatEnumEncoderName(enumTypeName)).Call(
							Id(encoderVariableName),
							nameFormatter(field).Index(Id("i")),
						)
						forBody.If(
							Err().Op("!=").Nil(),
						).Block(
							ReturnFunc(
								func(returnBody *Group) {
									if returnNilErr {
										returnBody.Nil()
									}
									returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
										Lit(exportedArgName),
										Qual(PkgAnchorGoErrors, "NewIndex").Call(
											Id("i"),
											Err(),
										),
									)
								},
							),
						)
					})
				})
			case *idltype.Vec:
				enumTypeName := field.Ty.(*idltype.Vec).Vec.(*idltype.Defined).Name
				body.BlockFunc(func(argBody *Group) {
					argBody.Err().Op(":=").Id(encoderVariableName).Dot("WriteLength").Call(
						Len(nameFormatter(field)),
					)
					argBody.If(
						Err().Op("!=").Nil(),
					).Block(
						ReturnFunc(
							func(returnBody *Group) {
								if returnNilErr {
									returnBody.Nil()
								}
								returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
									Lit(exportedArgName),
									Qual("fmt", "Errorf").Call(
										Lit("error while writing vector length: %w"),
										Err(),
									),
								)
							},
						),
					)
					argBody.For(
						Id("i").Op(":=").Lit(0),
						Id("i").Op("<").Len(nameFormatter(field)),
						Id("i").Op("++"),
					).BlockFunc(func(forBody *Group) {
						forBody.Err().Op(":=").Id(formatEnumEncoderName(enumTypeName)).Call(
							Id(encoderVariableName),
							nameFormatter(field).Index(Id("i")),
						)
						forBody.If(
							Err().Op("!=").Nil(),
						).Block(
							ReturnFunc(
								func(returnBody *Group) {
									if returnNilErr {
										returnBody.Nil()
									}
									returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
										Lit(exportedArgName),
										Qual(PkgAnchorGoErrors, "NewIndex").Call(
											Id("i"),
											Err(),
										),
									)
								},
							),
						)
					})
				})
			}
		} else {
			if IsOption(field.Ty) || IsCOption(field.Ty) {
				var optionalityWriterName string
				if IsOption(field.Ty) {
					optionalityWriterName = "WriteOption"
				} else {
					optionalityWriterName = "WriteCOption"
				}
				if checkNil {
					body.BlockFunc(func(optGroup *Group) {
						// if nil:
						optGroup.If(nameFormatter(field).Op("==").Nil()).Block(
							Err().Op("=").Id(encoderVariableName).Dot(optionalityWriterName).Call(False()),
							If(Err().Op("!=").Nil()).Block(
								ReturnFunc(
									func(returnBody *Group) {
										if returnNilErr {
											returnBody.Nil()
										}
										returnBody.Qual(PkgAnchorGoErrors, "NewOption").Call(
											Lit(exportedArgName),
											Qual("fmt", "Errorf").Call(
												Lit("error while encoding optionality: %w"),
												Err(),
											),
										)
									},
								),
							),
						).Else().Block(
							Err().Op("=").Id(encoderVariableName).Dot(optionalityWriterName).Call(True()),
							If(Err().Op("!=").Nil()).Block(
								ReturnFunc(
									func(returnBody *Group) {
										if returnNilErr {
											returnBody.Nil()
										}
										returnBody.Qual(PkgAnchorGoErrors, "NewOption").Call(
											Lit(exportedArgName),
											Qual("fmt", "Errorf").Call(
												Lit("error while encoding optionality: %w"),
												Err(),
											),
										)
									},
								),
							),
							Err().Op("=").Id(encoderVariableName).Dot("Encode").Call(nameFormatter(field)),
							If(Err().Op("!=").Nil()).Block(
								ReturnFunc(
									func(returnBody *Group) {
										if returnNilErr {
											returnBody.Nil()
										}
										returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
											Lit(exportedArgName),
											Err(),
										)
									},
								),
							),
						)
					})
				} else {
					body.BlockFunc(func(optGroup *Group) {
						// TODO: make optional fields of accounts a pointer.
						// Write as if not nil:
						optGroup.Err().Op("=").Id(encoderVariableName).Dot(optionalityWriterName).Call(True())
						optGroup.If(Err().Op("!=").Nil()).Block(
							ReturnFunc(
								func(returnBody *Group) {
									if returnNilErr {
										returnBody.Nil()
									}
									returnBody.Qual(PkgAnchorGoErrors, "NewOption").Call(
										Lit(exportedArgName),
										Qual("fmt", "Errorf").Call(
											Lit("error while encoding optionality: %w"),
											Err(),
										),
									)
								},
							),
						)
						optGroup.Err().Op("=").Id(encoderVariableName).Dot("Encode").Call(nameFormatter(field))
						optGroup.If(Err().Op("!=").Nil()).Block(
							ReturnFunc(
								func(returnBody *Group) {
									if returnNilErr {
										returnBody.Nil()
									}
									returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
										Lit(exportedArgName),
										Err(),
									)
								},
							),
						)
					})
				}
			} else {
				body.Err().Op("=").Id(encoderVariableName).Dot("Encode").Call(nameFormatter(field))
				body.If(Err().Op("!=").Nil()).Block(
					ReturnFunc(
						func(returnBody *Group) {
							if returnNilErr {
								returnBody.Nil()
							}
							returnBody.Qual(PkgAnchorGoErrors, "NewField").Call(
								Lit(exportedArgName),
								Err(),
							)
						},
					),
				)
			}
		}
	}
}
