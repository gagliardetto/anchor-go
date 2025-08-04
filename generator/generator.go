package generator

import (
	"fmt"
	"log/slog"

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
	Logger      *slog.Logger
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
		if len(g.idl.Docs) > 0 {
			g.log("Generating docs")
			file, err := g.genfile_doc()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Docs generated")
		}
		if len(g.idl.Accounts) > 0 {
			g.log("Generating accounts")
			file, err := g.genfile_accounts()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Accounts generated")
		}
		if len(g.idl.Events) > 0 {
			g.log("Generating events")
			file, err := g.genfile_events()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Events generated")
		}
		{
			g.log("Generating types")
			file, err := g.genfile_types()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Types generated")
		}
		{
			g.log("Generating discriminators")
			file, err := g.gen_discriminators()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Discriminators generated")
		}
		{
			g.log("Generating fetchers")
			file, err := g.gen_fetchers()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Fetchers generated")
		}
		{
			g.log("Generating errors")
			file, err := g.gen_errors()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Errors generated")
		}
		{
			g.log("Generating constants")
			file, err := g.gen_constants()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Constants generated")
		}
		{
			g.log("Generating tests")
			file, err := g.gen_tests()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Tests generated")
		}
		{
			g.log("Generating instructions")
			file, err := g.gen_instructions()
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Instructions generated")
		}
		if g.options.ProgramId != nil {
			g.log("Generating program ID")
			file, err := g.genfile_programID(*g.options.ProgramId)
			if err != nil {
				return nil, err
			}
			output.Files = append(output.Files, file)
			g.logSuccess("Program ID generated")
		}
		if !g.options.SkipGoMod {
			g.log("Generating go.mod")
			goMod, err := g.gen_gomod()
			if err != nil {
				return nil, err
			}
			output.GoMod = goMod
			g.logSuccess("go.mod generated")
		}
		g.logSuccess("Client generated successfully")
	}

	return output, nil
}

func (g *Generator) log(msg string, args ...any) {
	if g.options.Logger != nil {
		g.options.Logger.Info("[INFO] "+msg, args...)
	}
}

func (g *Generator) logSuccess(msg string, args ...any) {
	if g.options.Logger != nil {
		g.options.Logger.Info("[OK] "+msg, args...)
	}
}
