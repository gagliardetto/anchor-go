package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

type getterFieldSpec struct {
	FieldName string
	Ty        idltype.IdlType
	Optional  bool
}

var reservedGetterMethodNames = map[string]struct{}{
	"GetDiscriminator":     {},
	"GetAccountKeys":       {},
	"Unmarshal":            {},
	"UnmarshalWithDecoder": {},
	"MarshalWithEncoder":   {},
}

func getterSpecsFromDefinedFields(fields idl.IdlDefinedFields) []getterFieldSpec {
	switch typedFields := fields.(type) {
	case idl.IdlDefinedFieldsNamed:
		uniqueFieldNames := generateUniqueFieldNames(typedFields)
		specs := make([]getterFieldSpec, 0, len(typedFields))
		for _, field := range typedFields {
			specs = append(specs, getterFieldSpec{
				FieldName: uniqueFieldNames[field.Name],
				Ty:        field.Ty,
				Optional:  IsOption(field.Ty) || IsCOption(field.Ty),
			})
		}
		return specs
	case idl.IdlDefinedFieldsTuple:
		specs := make([]getterFieldSpec, 0, len(typedFields))
		for i, field := range typedFields {
			specs = append(specs, getterFieldSpec{
				FieldName: FormatTupleItemName(i),
				Ty:        field,
				Optional:  IsOption(field) || IsCOption(field),
			})
		}
		return specs
	default:
		return nil
	}
}

func getterSpecsFromInstruction(instruction idl.IdlInstruction) []getterFieldSpec {
	specs := make([]getterFieldSpec, 0, len(instruction.Args)+(len(instruction.Accounts)*4))
	for _, arg := range instruction.Args {
		specs = append(specs, getterFieldSpec{
			FieldName: tools.ToCamelUpper(arg.Name),
			Ty:        arg.Ty,
			Optional:  IsOption(arg.Ty) || IsCOption(arg.Ty),
		})
	}

	for _, account := range instruction.Accounts {
		acc, ok := account.(*idl.IdlInstructionAccount)
		if !ok {
			continue
		}

		fieldName := tools.ToCamelUpper(acc.Name)
		specs = append(specs, getterFieldSpec{
			FieldName: fieldName,
			Ty:        &idltype.Pubkey{},
		})

		if acc.Writable {
			specs = append(specs, getterFieldSpec{
				FieldName: fieldName + "Writable",
				Ty:        &idltype.Bool{},
			})
		}
		if acc.Signer {
			specs = append(specs, getterFieldSpec{
				FieldName: fieldName + "Signer",
				Ty:        &idltype.Bool{},
			})
		}
		if acc.Optional {
			specs = append(specs, getterFieldSpec{
				FieldName: fieldName + "Optional",
				Ty:        &idltype.Bool{},
			})
		}
	}

	return specs
}

func genGetterMethods(receiverTypeName string, specs []getterFieldSpec) Code {
	code := Empty()
	for _, spec := range specs {
		getterName := getterMethodName(spec.FieldName)

		code.Line().Line()
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id(getterName).
			Params().
			Add(genTypeName(spec.Ty)).
			BlockFunc(func(block *Group) {
				if spec.Optional {
					block.If(Id("obj").Op("==").Nil().Op("||").Id("obj").Dot(spec.FieldName).Op("==").Nil()).BlockFunc(func(ifBlock *Group) {
						returnZeroValue(ifBlock, spec.Ty)
					})
					block.Return(Op("*").Id("obj").Dot(spec.FieldName))
					return
				}

				block.If(Id("obj").Op("==").Nil()).BlockFunc(func(ifBlock *Group) {
					returnZeroValue(ifBlock, spec.Ty)
				})
				block.Return(Id("obj").Dot(spec.FieldName))
			})

		if !spec.Optional {
			continue
		}

		code.Line().Line()
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id(getterPtrMethodName(spec.FieldName)).
			Params().
			Params(Op("*").Add(genTypeName(spec.Ty))).
			BlockFunc(func(block *Group) {
				block.If(Id("obj").Op("==").Nil()).Block(
					Return(Nil()),
				)
				block.Return(Id("obj").Dot(spec.FieldName))
			})

		code.Line().Line()
		code.Func().Params(Id("obj").Op("*").Id(receiverTypeName)).Id(hasMethodName(spec.FieldName)).
			Params().
			Bool().
			Block(
				Return(Id("obj").Op("!=").Nil().Op("&&").Id("obj").Dot(spec.FieldName).Op("!=").Nil()),
			)
	}

	return code
}

func getterMethodName(fieldName string) string {
	methodName := "Get" + tools.ToCamelUpper(fieldName)
	if _, exists := reservedGetterMethodNames[methodName]; exists {
		return methodName + "Field"
	}
	return methodName
}

func getterPtrMethodName(fieldName string) string {
	return getterMethodName(fieldName) + "Ptr"
}

func hasMethodName(fieldName string) string {
	return "Has" + tools.ToCamelUpper(fieldName)
}

func returnZeroValue(block *Group, fieldType idltype.IdlType) {
	block.Var().Id("zero").Add(genTypeName(fieldType))
	block.Return(Id("zero"))
}
