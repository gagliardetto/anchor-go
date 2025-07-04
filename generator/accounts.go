package generator

import (
	"fmt"
	"strconv"

	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

func (g *Generator) genfile_accounts() (*OutputFile, error) {
	file := NewFile(g.options.Package)
	file.HeaderComment("Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.")
	file.HeaderComment("This file contains parsers for the accounts defined in the IDL.")

	names := []string{}
	{
		for _, acc := range g.idl.Accounts {
			names = append(names, tools.ToCamelUpper(acc.Name))
		}
	}
	{
		code, err := g.gen_accountParser(names)
		if err != nil {
			return nil, fmt.Errorf("error generating account parser: %w", err)
		}
		file.Add(code)
	}

	return &OutputFile{
		Name: "accounts.go",
		File: file,
	}, nil
}

func (g *Generator) gen_accountParser(accountNames []string) (Code, error) {
	code := Empty()
	{
		code.Func().Id("ParseAnyAccount").
			Params(Id("accountData").Index().Byte()).
			Params(Any(), Error()).
			BlockFunc(func(block *Group) {
				block.Id("decoder").Op(":=").Qual(PkgBinary, "NewBorshDecoder").Call(Id("accountData"))
				block.List(Id("discriminator"), Err()).Op(":=").Id("decoder").Dot("ReadDiscriminator").Call()

				block.If(Err().Op("!=").Nil()).Block(
					Return(
						Nil(),
						Qual("fmt", "Errorf").Call(Lit("failed to peek account discriminator: %w"), Err()),
					),
				)

				block.Switch(Id("discriminator")).BlockFunc(func(switchBlock *Group) {
					for _, name := range accountNames {
						switchBlock.Case(Id(FormatAccountDiscriminatorName(name))).Block(
							Id("value").Op(":=").New(Id(name)),
							Err().Op(":=").Id("value").Dot("UnmarshalWithDecoder").Call(Id("decoder")),
							If(Err().Op("!=").Nil()).Block(
								Return(
									Nil(),
									Qual("fmt", "Errorf").Call(Lit("failed to unmarshal account as "+name+": %w"), Err()),
								),
							),
							Return(Id("value"), Nil()),
						)
					}
					switchBlock.Default().Block(
						Return(Nil(), Qual("fmt", "Errorf").Call(Lit("unknown discriminator: %s"), Qual(PkgBinary, "FormatDiscriminator").Call(Id("discriminator")))),
					)
				})
			})
	}
	{
		code.Line().Line()
		// for each account, generate a function to parse it:
		for _, name := range accountNames {
			discriminatorName := FormatAccountDiscriminatorName(name)

			code.Func().Id("ParseAccount_"+name).
				Params(Id("accountData").Index().Byte()).
				Params(Op("*").Id(name), Error()).
				BlockFunc(func(block *Group) {
					block.Id("decoder").Op(":=").Qual(PkgBinary, "NewBorshDecoder").Call(Id("accountData"))
					block.List(Id("discriminator"), Err()).Op(":=").Id("decoder").Dot("ReadDiscriminator").Call()

					block.If(Err().Op("!=").Nil()).Block(
						Return(
							Nil(),
							Qual("fmt", "Errorf").Call(Lit("failed to peek discriminator: %w"), Err()),
						),
					)

					block.If(Id("discriminator").Op("!=").Id(discriminatorName)).Block(
						Return(Nil(), Qual("fmt", "Errorf").Call(Lit("expected discriminator %v, got %s"), Id(discriminatorName), Qual(PkgBinary, "FormatDiscriminator").Call(Id("discriminator")))),
					)

					block.Id("acc").Op(":=").New(Id(name))
					block.Err().Op("=").Id("acc").Dot("UnmarshalWithDecoder").Call(Id("decoder"))

					block.If(Err().Op("!=").Nil()).Block(
						Return(
							Nil(),
							Qual("fmt", "Errorf").Call(Lit("failed to unmarshal account of type "+name+": %w"), Err()),
						),
					)

					block.Return(Id("acc"), Nil())
				})
			code.Line().Line()
		}
	}
	return code, nil
}

func (g *Generator) gen_IDLTypeDefTyStruct(
	name string,
	docs []string,
	typ idl.IdlTypeDefTyStruct,
	withDiscriminator bool,
) (Code, error) {
	st := newStatement()

	exportedAccountName := tools.ToCamelUpper(name)
	{
		// Declare the struct:
		code := Empty()
		addComments(code, docs)
		code.Type().Id(exportedAccountName).StructFunc(func(fieldsGroup *Group) {
			switch fields := typ.Fields.(type) {
			case idl.IdlDefinedFieldsNamed:
				for fieldIndex, field := range fields {

					// Add docs for the field:
					for docIndex, doc := range field.Docs {
						if docIndex == 0 && fieldIndex > 0 {
							fieldsGroup.Line()
						}
						fieldsGroup.Comment(doc)
					}
					// fieldsGroup.Line()
					optionality := IsOption(field.Ty) || IsCOption(field.Ty)

					// TODO: optionality for complex enums is a nil interface.
					fieldsGroup.Add(genField(field, optionality)).
						Add(func() Code {
							tagMap := map[string]string{}
							if IsOption(field.Ty) {
								tagMap["bin"] = "optional"
							}
							if IsCOption(field.Ty) {
								tagMap["bin"] = "coption"
							}
							// add json tag:
							tagMap["json"] = tools.ToCamelLower(field.Name) + func() string {
								if optionality {
									return ",omitempty"
								}
								return ""
							}()
							return Tag(tagMap)
						}())
				}
			case idl.IdlDefinedFieldsTuple:
				// panic(fmt.Errorf("tuple fields not supported: %s", spew.Sdump(fields)))
				for fieldIndex, field := range fields {

					fieldsGroup.Line()
					optionality := IsOption(field) || IsCOption(field)

					fieldsGroup.Add(genFieldNamed(
						FormatTupleItemName(fieldIndex),
						field,
						optionality,
					)).
						Add(func() Code {
							tagMap := map[string]string{}
							if IsOption(field) {
								tagMap["bin"] = "optional"
							}
							if IsCOption(field) {
								tagMap["bin"] = "coption"
							}
							// add json tag:
							tagMap["json"] = tools.ToCamelLower(FormatTupleItemName(fieldIndex)) + func() string {
								if optionality {
									return ",omitempty"
								}
								return ""
							}()
							return Tag(tagMap)
						}())
				}

			case nil:
				// No fields, just an empty struct.
				// TODO: should we panic here?
			default:
				panic(fmt.Errorf("unknown fields type: %T", typ.Fields))
			}
		})
		st.Add(code.Line())
	}
	{
		// Declare the decoder/encoder methods:
		code := Empty()

		{
			discriminatorName := FormatAccountDiscriminatorName(exportedAccountName)

			// Declare MarshalWithEncoder:
			// TODO:
			code.Line().Line().Add(
				gen_MarshalWithEncoder_struct(
					g.idl,
					withDiscriminator,
					exportedAccountName,
					discriminatorName,
					typ.Fields,
					true,
				))

			// Declare UnmarshalWithDecoder
			code.Line().Line().Add(
				gen_UnmarshalWithDecoder_struct(
					g.idl,
					withDiscriminator,
					exportedAccountName,
					discriminatorName,
					typ.Fields,
				))
		}
		st.Add(code.Line().Line())
	}
	return st, nil
}

func genField(field idl.IdlField, pointer bool) Code {
	return genFieldNamed(field.Name, field.Ty, pointer)
}

func genFieldNamed(name string, typ idltype.IdlType, pointer bool) Code {
	st := newStatement()
	st.Id(tools.ToCamelUpper(name)).
		Add(func() Code {
			if isComplexEnum(typ) {
				return nil
			}
			if pointer {
				return Op("*")
			}
			return nil
		}()).
		Add(genTypeName(typ))
	return st
}

func genTypeName(idlTypeEnv idltype.IdlType) Code {
	st := newStatement()
	switch {
	case IsIDLTypeKind(idlTypeEnv):
		{
			str := idlTypeEnv
			st.Add(IDLTypeKind_ToTypeDeclCode(str))
		}
	case IsOption(idlTypeEnv):
		{
			opt := idlTypeEnv.(*idltype.Option)
			// TODO: optional = pointer? or that's determined upstream?
			st.Add(genTypeName(opt.Option))
		}
	case IsCOption(idlTypeEnv):
		{
			copt := idlTypeEnv.(*idltype.COption)
			st.Add(genTypeName(copt.COption))
		}
	case IsVec(idlTypeEnv):
		{
			vec := idlTypeEnv.(*idltype.Vec)
			st.Index().Add(genTypeName(vec.Vec))
		}
	case IsDefined(idlTypeEnv):
		{
			def := idlTypeEnv.(*idltype.Defined)
			st.Add(Id(tools.ToCamelUpper(def.Name)))
		}
	case IsArray(idlTypeEnv):
		{
			arr := idlTypeEnv.(*idltype.Array)
			{
				switch size := arr.Size.(type) {
				case *idltype.IdlArrayLenGeneric:
					panic(fmt.Sprintf("generic array length not supported: %s", spew.Sdump(size)))
				case *idltype.IdlArrayLenValue:
					if size.Value < 0 {
						panic(fmt.Sprintf("expected positive integer, got %d", size.Value))
					}
					st.Index(Id(strconv.Itoa(int(size.Value)))).Add(genTypeName(arr.Type))
				}
			}
		}
	default:
		panic("unhandled type: " + spew.Sdump(idlTypeEnv))
	}
	return st
}

func IDLTypeKind_ToTypeDeclCode(ts idltype.IdlType) *Statement {
	stat := newStatement()
	switch ts.(type) {
	case *idltype.Bool:
		stat.Bool()
	case *idltype.U8:
		stat.Uint8()
	case *idltype.I8:
		stat.Int8()
	case *idltype.U16:
		// TODO: some types have their implementation in github.com/gagliardetto/binary
		stat.Uint16()
	case *idltype.I16:
		stat.Int16()
	case *idltype.U32:
		stat.Uint32()
	case *idltype.I32:
		stat.Int32()
	case *idltype.F32:
		stat.Float32()
	case *idltype.U64:
		stat.Uint64()
	case *idltype.I64:
		stat.Int64()
	case *idltype.F64:
		stat.Float64()
	case *idltype.U128:
		stat.Qual(PkgBinary, "Uint128")
	case *idltype.I128:
		stat.Qual(PkgBinary, "Int128")
	case *idltype.Bytes:
		stat.Index().Byte()
	case *idltype.String:
		stat.String()
	case *idltype.Pubkey:
		stat.Qual(PkgSolanaGo, "PublicKey")

	default:
		panic(fmt.Sprintf("unhandled type: %s", spew.Sdump(ts)))
	}

	return stat
}
