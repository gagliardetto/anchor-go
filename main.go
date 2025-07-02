package main

import (
	"flag"
	"fmt"
	"go/token"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/gagliardetto/anchor-go/generator"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/tools"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

const defaultProgramName = "myprogram"

func main() {
	if askingForVersion() {
		printVersion()
		return
	}
	var outputDir string
	var programName string
	var modPath string
	var pathToIdl string
	var programIDOverride solana.PublicKey
	flag.Var(&programIDOverride, "program-id", "Program ID to use in the generated code (optional)")
	flag.StringVar(&outputDir, "output", "", "Directory to write the generated code to")
	flag.StringVar(&programName, "name", defaultProgramName, "Name of the program for the generated code")
	flag.StringVar(&modPath, "mod-path", "", "Module path for the generated code (optional)")
	flag.StringVar(&pathToIdl, "idl", "", "Path to the IDL file (required)")
	var skipGoMod bool
	flag.BoolVar(&skipGoMod, "no-go-mod", false, "Skip generating the go.mod file (useful for testing)")
	flag.Parse()
	if pathToIdl == "" {
		panic("Please provide the path to the IDL file using the -idl flag")
	}
	if outputDir == "" {
		panic("Please provide the output directory using the -output flag")
	}

	if modPath == "" {
		modPath = path.Join("github.com", "gagliardetto", "anchor-go", "generated")
		slog.Info("Using default module path", "modPath", modPath)
	} else {
		slog.Info("Using provided module path", "modPath", modPath)
	}
	if err := os.MkdirAll(outputDir, 0o777); err != nil {
		panic(fmt.Errorf("Failed to create output directory: %w", err))
	}
	slog.Info("Starting code generation",
		"outputDir", outputDir,
		"modPath", modPath,
		"pathToIdl", pathToIdl,
		"programID", func() string {
			if programIDOverride.IsZero() {
				return "not provided"
			}
			return programIDOverride.String()
		}(),
	)

	options := generator.GeneratorOptions{
		OutputDir:   outputDir,
		Package:     programName,
		ProgramName: programName,
		ModPath:     modPath,
		SkipGoMod:   skipGoMod,
	}
	if !programIDOverride.IsZero() {
		options.ProgramId = &programIDOverride
		slog.Info("Using provided program ID", "programID", programIDOverride.String())
	}
	parsedIdl, err := idl.ParseFromFilepath(pathToIdl)
	if err != nil {
		panic(err)
	}
	if parsedIdl == nil {
		panic("Parsed IDL is nil, please check the IDL file path and format.")
	}
	if err := parsedIdl.Validate(); err != nil {
		panic(fmt.Errorf("Invalid IDL: %w", err))
	}
	{
		{
			if parsedIdl.Address != nil && !parsedIdl.Address.IsZero() && options.ProgramId == nil {
				// If the IDL has an address, use it as the program ID:
				slog.Info("Using IDL address as program ID", "address", parsedIdl.Address.String())
				options.ProgramId = parsedIdl.Address
			}
		}
		parsedIdl.Metadata.Name = bin.ToSnakeForSighash(parsedIdl.Metadata.Name)
		{
			// check that the name is not a reserved keyword:
			if parsedIdl.Metadata.Name != "" {
				if tools.IsReservedKeyword(parsedIdl.Metadata.Name) {
					slog.Warn("The IDL metadata.name is a reserved Go keyword: adding a suffix to avoid conflicts.",
						"name", parsedIdl.Metadata.Name,
						"reservedKeyword", token.Lookup(parsedIdl.Metadata.Name).String(),
					)
					// Add a suffix to the name to avoid conflicts with Go reserved keywords:
					parsedIdl.Metadata.Name += "_program"
				}
				if !tools.IsValidIdent(parsedIdl.Metadata.Name) {
					// add a prefix to the name to avoid conflicts with Go reserved keywords:
					parsedIdl.Metadata.Name = "my_" + parsedIdl.Metadata.Name
				}
			}
			// if begins with
		}
		if programName == "" && parsedIdl.Metadata.Name != "" {
			panic("Please provide a package name using the -name flag, or ensure the IDL has a valid metadata.name field.")
		}
		if programName == defaultProgramName && parsedIdl.Metadata.Name != "" {
			cleanedName := bin.ToSnakeForSighash(parsedIdl.Metadata.Name)
			options.Package = cleanedName
			options.ProgramName = cleanedName
			slog.Info("Using IDL metadata.name as package name", "packageName", cleanedName)
		}

		slog.Info("Parsed IDL successfully",
			"version", parsedIdl.Metadata.Version,
			"name", parsedIdl.Metadata.Name,
			"address", parsedIdl.Address,
			"programId", func() string {
				if parsedIdl.Address.IsZero() {
					return "not provided"
				}
				return parsedIdl.Address.String()
			}(),
			"instructionsCount", len(parsedIdl.Instructions),
			"accountsCount", len(parsedIdl.Accounts),
			"eventsCount", len(parsedIdl.Events),
			"typesCount", len(parsedIdl.Types),
			"constantsCount", len(parsedIdl.Constants),
			"errorsCount", len(parsedIdl.Errors),
		)
	}
	gen := generator.NewGenerator(parsedIdl, &options)
	generatedFiles, err := gen.Generate()
	if err != nil {
		panic(err)
	}

	if !skipGoMod {
		goModFilepath := path.Join(options.OutputDir, "go.mod")
		slog.Info("Writing go.mod file",
			"filepath", goModFilepath,
			"modPath", options.ModPath,
		)

		err = os.WriteFile(goModFilepath, []byte(generatedFiles.GoMod), 0o777)
		if err != nil {
			panic(err)
		}
	}
	{
		for _, file := range generatedFiles.Files {
			{
				// Save assets:
				assetFilename := file.Name
				assetFilepath := path.Join(options.OutputDir, assetFilename)

				// Create file:
				goFile, err := os.Create(assetFilepath)
				if err != nil {
					panic(err)
				}
				defer goFile.Close()

				slog.Info("Writing file",
					"filepath", assetFilepath,
					"name", file.Name,
					"modPath", options.ModPath,
				)
				err = file.File.Render(goFile)
				if err != nil {
					panic(err)
				}
			}
		}
		executeCmd(outputDir, "go", "mod", "tidy")
		executeCmd(outputDir, "go", "fmt")
		executeCmd(outputDir, "go", "build", "-o", "/dev/null") // Just to ensure everything compiles.
		slog.Info("Generation completed successfully",
			"outputDir", options.OutputDir,
			"modPath", options.ModPath,
			"package", options.Package,
			"programName", options.ProgramName,
		)
	}
}

func executeCmd(dir string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
