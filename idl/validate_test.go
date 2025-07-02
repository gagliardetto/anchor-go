package idl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gagliardetto/anchor-go/internal"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	list, err := filepath.Glob("test-idls/2025-may-19/*.json")
	require.NoError(t, err)

	numSkipped := 0

	for _, path := range list {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if internal.IsKnownBrokenIdl(filepath.Base(path)) {
				t.Skip("Skipping known broken IDL", filepath.Base(path))
			}
			src, err := os.ReadFile(path)
			require.NoError(t, err, string(src))

			if IsOldIdl(src) {
				numSkipped++
				// Old IDL format; skip parsing
				fmt.Println("	Skipping old IDL format", path)
				t.Skipf("Old IDL format for %s; skipping", path)
				return
			}

			schema := new(Idl)
			err = json.Unmarshal(src, &schema)
			require.NoError(t, err)

			validationErrs := ValidateIDL(schema)
			if validationErrs != nil {
				t.Error(validationErrs)
			}
		})
	}
}
