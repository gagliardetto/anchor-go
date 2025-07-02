package generator

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/solana-go"
)

var Debug = false // Set to true to enable debug logging.

type Generator struct {
	options *GeneratorOptions
	idl     *idl.Idl
}

type GeneratorOptions struct {
	OutputDir   string            // Directory to write the generated code to.
	Package     string            // Package name for the generated code.
	ModPath     string            // Module path for the generated code. E.g. "github.com/gagliardetto/mysolana-program-go"
	ProgramId   *solana.PublicKey // Program ID to use in the generated code.
	ProgramName string            // Name of the program for the generated code.
	SkipGoMod   bool              // If true, skip generating the go.mod file.
}

func NewGenerator(idl *idl.Idl, options *GeneratorOptions) *Generator {
	return &Generator{
		idl:     idl,
		options: options,
	}
}

type OutputFile struct {
	Name string // Name of the output file.
	File *File
}

type Output struct {
	Files []*OutputFile // List of output files to be generated.
	GoMod []byte        // Go module file content.
}

func (g *Generator) Generate() (*Output, error) {
	if g.idl == nil {
		return nil, fmt.Errorf("IDL is nil, cannot generate code")
	}
	if g.options == nil {
		g.options = &GeneratorOptions{
			OutputDir:   "generated",
			Package:     "idlclient",
			ModPath:     "github.com/gagliardetto/anchor-go/idlclient",
			ProgramId:   nil,
			ProgramName: "myprogram",
		}
	}
	if err := g.idl.Validate(); err != nil {
		return nil, fmt.Errorf("invalid IDL: %w", err)
	}
	output := &Output{
		Files: make([]*OutputFile, 0),
	}

	{
		// Register complex enums.
		{
			// register complex enums:
			// TODO: .types is the only place where we can find complex enums? (or enums in general?)
			for _, typ := range g.idl.Types {
				registerComplexEnums(typ)
			}
		}
		{
			file, err := g.genfile_doc()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		if g.idl.Accounts != nil && len(g.idl.Accounts) > 0 {
			file, err := g.genfile_accounts()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		if g.idl.Events != nil && len(g.idl.Events) > 0 {
			file, err := g.genfile_events()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.genfile_types()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_discriminators()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_fetchers()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_errors()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_constants()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_tests()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		{
			file, err := g.gen_instructions()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		if g.options.ProgramId != nil {
			file, err := g.genfile_programID(*g.options.ProgramId)
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
		}
		if !g.options.SkipGoMod {
			goMod, err := g.gen_gomod()
			if err != nil {
				return nil, err
			}
			output.GoMod = goMod
		}
	}

	return output, nil
}
