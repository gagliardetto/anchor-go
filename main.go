package main

import (
	"encoding/json"
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
		// "idl_files/pyth.json",
		// "idl_files/multisig.json",
		// "idl_files/misc.json",
		// "idl_files/lockup.json",
		// "idl_files/ido_pool.json",
		// "idl_files/events.json",
		// "idl_files/escrow.json",
		// "idl_files/errors.json",
		// "idl_files/composite.json",
		"idl_files/chat.json",
		// "idl_files/cashiers_check.json",
		// "idl_files/counter_auth.json",
		// "idl_files/counter.json",
	}
	for _, idlFilepath := range filenames {
		Ln(LimeBG(idlFilepath))
		// idlFilepath := "/home/withparty/go/src/github.com/project-serum/anchor/examples/escrow/target/idl/escrow.json"
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

func GenerateClient(idl IDL) error {
	// TODO:
	// - validate IDL (???)
	// - create new go file
	// - add instructions (aka methods)

	file := NewGoFile(idl.Name, true)

	{
		code := Type().Id("Client").Struct(
			Id("rpcURL").String(),
			Id("rpcClient").Qual("github.com/ybbus/jsonrpc", "RPCClient"),
		)
		file.Add(code.Line())
	}

	for _, typ := range idl.Types {
		switch typ.Type.Kind {
		case IdlTypeDefTyKindStruct:
			code := Empty()
			code.Type().Id(typ.Name).StructFunc(func(fieldsGroup *Group) {
				for _, field := range *typ.Type.Fields {
					fieldsGroup.Id(ToCamel(field.Name)).Add(
						DoGroup(func(fieldTypeGroup *Group) {
							if field.Type.IsString() {
								fieldTypeGroup.Add(typeStringToType(field.Type.GetString()))
							}

							if field.Type.IsArray() {
								arr := field.Type.GetArray()
								_ = arr

								if arr.Thing.IsString() {
									fieldTypeGroup.Index()
									fieldTypeGroup.Add(typeStringToType(arr.Thing.GetString()))
								}
							}
						}),
					)
				}
			})

			file.Add(code.Line())
		case IdlTypeDefTyKindEnum:
			panic("not implemented")
		}

	}

	for _, instruction := range idl.Instructions {
		methodExportedName := ToCamel(instruction.Name)

		code := Empty()
		code.Commentf(
			"%s method sends the `%s` instruction.",
			methodExportedName,
			instruction.Name,
		).Line()

		code.Func().Params(Id("cl").Op("*").Id("Client")).Id(methodExportedName).
			Params(
				ListFunc(func(st *Group) {
					// Parameters:
					st.Id("ctx").Qual("context", "Context")

					for _, arg := range instruction.Args {
						st.Id(arg.Name).Id(string(arg.Type.GetString()))
						// TODO: determine the right type for the arg.
					}

				}),
			).
			Params(
				ListFunc(func(st *Group) {
					// Results:
					st.Err().Error()
				}),
			).
			BlockFunc(func(gr *Group) {
				// Body:
				gr.Id("params").Op(":=").Index().Interface().Block(
					DoGroup(
						func(paramsGroup *Group) {
							for _, arg := range instruction.Args {
								paramsGroup.Id(arg.Name).Op(",")
							}
						},
					),
				)

				var populateIntoName string
				populateIntoName = "&out"
				_ = populateIntoName
				gr.Err().Op("=").Id("cl").Dot("rpcClient").Dot("CallFor").CallFunc(func(callGr *Group) {
					// callGr.Id(populateIntoName)
					callGr.Nil()
					callGr.Lit(instruction.Name)
					callGr.Id("params")
				})

				gr.Return()
			})
		file.Add(code.Line())
	}

	{
		err := file.Render(os.Stdout)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
