package main

import (
	"encoding/json"
	"testing"

	. "github.com/dave/jennifer/jen"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_genTypeName(t *testing.T) {
	type jsonToSource struct {
		from     string
		expected string
	}

	tests := []jsonToSource{
		//
		{
			`{"type": "pubkey"}`,
			"var thing solanago.PublicKey",
		},
		{
			`{"type": "bool"}`,
			"var thing bool",
		},
		{
			`{"type": "u8"}`,
			"var thing uint8",
		},
		{
			`{"type": "i8"}`,
			"var thing int8",
		},
		{
			`{"type": "u16"}`,
			"var thing uint16",
		},
		{
			`{"type": "i16"}`,
			"var thing int16",
		},
		{
			`{"type": "u32"}`,
			"var thing uint32",
		},
		{
			`{"type": "i32"}`,
			"var thing int32",
		},
		{
			`{"type": "u64"}`,
			"var thing uint64",
		},
		{
			`{"type": "i64"}`,
			"var thing int64",
		},
		{
			`{"type": "u128"}`,
			"var thing binary.Uint128",
		},
		{
			`{"type": "i128"}`,
			"var thing binary.Int128",
		},
		{
			// TODO: is this also OK as []byte ???
			`{"type": "bytes"}`,
			"var thing []byte",
		},
		{
			`{"type": "string"}`,
			"var thing string",
		},
		{
			`{"type": "pubkey"}`,
			"var thing solanago.PublicKey",
		},

		// "defined"
		{
			`{"type": {"defined":{"name": "Foo"}}}`,
			"var thing Foo",
		},
		{
			`{"type": {"defined":{"name": "bar"}}}`,
			"var thing bar",
		},

		// "array":
		{
			`{"type": {"array":["u8",280]}}`,
			"var thing [280]uint8",
		},
		{
			`{"type": {"array":[{"defined":{ "name": "Message" }},33607]}}`,
			"var thing [33607]Message",
		},
		{
			`{"type": {"array":[{"array":["u8",280]},33607]}}`,
			"var thing [33607][280]uint8",
		},
		{
			`{"type": {"array":[{"array":[{"defined": { "name": "Message" }},123]},33607]}}`,
			"var thing [33607][123]Message",
		},

		// "vec":
		{
			`{"type": {"vec": "pubkey"}}`,
			"var thing []solanago.PublicKey",
		},
		{
			`{"type": {"vec": {"defined": { "name": "TransactionAccount" }}}}`,
			"var thing []TransactionAccount",
		},
		{
			`{"type": {"vec": "bool"}}`,
			"var thing []bool",
		},
		{
			`{"type": {"vec": {"array":[{"array":[{"defined":{ "name": "Message" }},123]},33607]}}}`,
			"var thing [][33607][123]Message",
		},

		// "option":
		{
			`{"type": {"option": "string"}}`,
			"var thing string",
		},
		{
			`{"type": {"option": {"vec": {"array":[{"array":[{"defined":{ "name": "Message" }},123]},33607]}}}}`,
			"var thing [][33607][123]Message",
		},
		{
			`{"type": {"option": {"defined": { "name": "TransactionAccount" }}}}`,
			"var thing TransactionAccount",
		},
	}
	{
		for _, scenario := range tests {
			var target IdlField
			err := json.Unmarshal([]byte(scenario.from), &target)
			if err != nil {
				panic(err)
			}
			code := Var().Id("thing").Add(genTypeName(target.Type))
			got := codeToString(code)
			require.Equal(t, scenario.expected, got)
		}
	}
}

func Test_genField(t *testing.T) {
	type jsonToSource struct {
		from     string
		expected string
	}

	tests := []jsonToSource{
		{
			`{"name":"space","type":"u64"}`,
			"var thing struct {\n	Space uint64\n}",
		},
		{
			`{"name":"space","type": {"option": {"vec": {"array":[{"array":[{"defined":{"name": "Message"}},123]},33607]}}}}`,
			"var thing struct {\n	Space [][33607][123]Message\n}",
		},
	}
	{
		for _, scenario := range tests {
			var target IdlField
			err := json.Unmarshal([]byte(scenario.from), &target)
			if err != nil {
				panic(err)
			}
			code := Var().Id("thing").Struct(
				genField(target, false),
			)
			got := codeToString(code)
			require.Equal(t, scenario.expected, got)
		}
	}
}

func Test_IdlAccountItemSlice_Walk(t *testing.T) {
	data := `[
		{
			"name": "authorityBefore",
			"signer": true
		},
		{
			"name": "marketGroup",
			"accounts": [
				{
					"name": "marketMarket",
					"writable": true
				},
				{
					"name": "foo",
					"writable": true
				},
				{
					"name": "subMarket",
					"accounts": [
						{
							"name": "subMarketMarket",
							"writable": true
						},
						{
							"name": "openOrders",
							"writable": true
						} 
					]
				}
			]
		},
		{
			"name": "authorityAfter",
			"writable": true
		}
	]`
	var target IdlAccountItemSlice
	err := json.Unmarshal([]byte(data), &target)
	if err != nil {
		panic(err)
	}

	spew.Dump(target)

	expectedGroups := []string{
		"instruction",
		"instruction/marketGroup",
		"instruction/marketGroup",
		"instruction/marketGroup/subMarket",
		"instruction/marketGroup/subMarket",
		"instruction",
	}
	gotGroups := []string{}

	expectedAccountNames := []string{
		"authorityBefore",
		"marketMarket",
		"foo",
		"subMarketMarket",
		"openOrders",
		"authorityAfter",
	}
	gotAccountNames := []string{}

	expectedIndexes := []int{0, 1, 2, 3, 4, 5}
	gotIndexes := []int{}

	instructionName := "instruction"
	target.Walk(instructionName, nil, nil, func(parentGroupPath string, index int, parentGroup *IdlAccounts, ia *IdlAccount) bool {
		gotGroups = append(gotGroups, parentGroupPath)
		gotAccountNames = append(gotAccountNames, ia.Name)
		gotIndexes = append(gotIndexes, index)
		return true
	})

	require.Equal(t, expectedGroups, gotGroups)
	require.Equal(t, expectedAccountNames, gotAccountNames)
	require.Equal(t, expectedIndexes, gotIndexes)
}

func TestFormatAccountAccessorName(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "Foo"))
		assert.Equal(t, "GetFooAccountAccount", formatAccountAccessorName("Get", "FooAccount"))
	})

	t.Run("remove config on", func(t *testing.T) {
		oldConf := GetConfig()
		defer func() {
			conf = oldConf
		}()
		conf = &Config{
			RemoveAccountSuffix: true,
		}

		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "Foo"))
		assert.Equal(t, "GetFooAccount", formatAccountAccessorName("Get", "FooAccount"))
	})
}
