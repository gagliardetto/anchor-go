package idl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdlParse(t *testing.T) {
	list, err := filepath.Glob("test-idls/2025-may-19/*.json")
	require.NoError(t, err)

	require.NotEmpty(t, list, "no test files found")
	numFound := len(list)
	numSkipped := 0
	numFailedParse := 0
	numFailedCompare := 0
	numSuccessParse := 0
	numSuccessCompare := 0

	for _, path := range list {
		t.Run(filepath.Base(path), func(t *testing.T) {
			fmt.Println(path)
			src, err := os.ReadFile(path)
			require.NoError(t, err, string(src))

			if IsOldIdl(src) {
				numSkipped++
				// Old IDL format; skip parsing
				fmt.Println("	Skipping old IDL format", path)
				t.Skipf("Old IDL format for %s; skipping", path)
				return
			}

			schema, err := Parse(src)
			if err != nil {
				numFailedParse++
			} else {
				numSuccessParse++
			}
			require.NoError(t, err)

			marshaled := mustAnyToJSON(t, schema)

			{
				if !jsoneq(t, (src), (marshaled)) {
					fmt.Println("	Failed to parse", path)
					fmt.Println("	Original:", string(src))
					fmt.Println("	Parsed:  ", string(marshaled))
					numFailedCompare++
					t.Fatalf("failed to parse %s", path)
				} else {
					numSuccessCompare++
				}
			}
		})
	}
	t.Logf("Parsed %d files, %d skipped (because old version), %d failed parse, %d failed compare, %d success parse, %d success compare",
		numFound,
		numSkipped,
		numFailedParse,
		numFailedCompare,
		numSuccessParse,
		numSuccessCompare,
	)
}

func jsoneq(t *testing.T, expected, actual []byte) bool {
	var expectedJSONAsInterface, actualJSONAsInterface any

	if err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface); err != nil {
		require.NoError(t, err, string(expected))
	}

	if err := json.Unmarshal([]byte(actual), &actualJSONAsInterface); err != nil {
		require.NoError(t, err, string(actual))
	}

	return assert.ObjectsAreEqual(expectedJSONAsInterface, actualJSONAsInterface)
}

func TestIdlParseOne(t *testing.T) {
	t.SkipNow()
	path := "test-idls/2025-may-19/13gDzEXCdocbj8iAiqrScGo47NiSuYENGsRqi3SEAwet.json"
	src, err := os.ReadFile(path)
	require.NoError(t, err, string(src))
	schema, err := Parse(src)
	require.NoError(t, err)
	marshaled := mustAnyToJSON(t, schema)
	{
		require.JSONEq(t, string(src), string(marshaled))
	}
}

func mustAnyToJSON(t require.TestingT, raw any) []byte {
	out, err := json.Marshal(raw)
	if err != nil {
		require.NoError(t, err, spew.Sdump(raw))
	}
	return out
}
