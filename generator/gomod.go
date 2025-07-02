package generator

import (
	"golang.org/x/mod/modfile"
)

// gen_gomod generates a `go.mod` file for the generated code, and writes
// it to the destination directory.
func (g *Generator) gen_gomod() ([]byte, error) {
	mdf := &modfile.File{}
	mdf.AddModuleStmt(g.options.ModPath)

	mdf.AddNewRequire("github.com/gagliardetto/solana-go", "v1.12.0", false)
	mdf.AddNewRequire("github.com/gagliardetto/anchor-go", "v0.3.2", false)
	mdf.AddNewRequire("github.com/gagliardetto/binary", "v0.8.0", false)
	mdf.AddNewRequire("github.com/gagliardetto/treeout", "v0.1.4", false)
	mdf.AddNewRequire("github.com/gagliardetto/gofuzz", "v1.2.2", false)
	mdf.AddNewRequire("github.com/stretchr/testify", "v1.10.0", false)
	mdf.AddNewRequire("github.com/davecgh/go-spew", "v1.1.1", false)

	// add replacement for "github.com/gagliardetto/anchor-go/errors" to ../../demo-anchor-go/errors
	// mdf.AddReplace("github.com/gagliardetto/anchor-go", "", "../../demo-anchor-go", "")
	mdf.Cleanup()

	return mdf.Format()
}
