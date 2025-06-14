package sighash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeForSighash(t *testing.T) {
	t.Run(
		"typescript",
		// "typescript package: https://www.npmjs.com/package/snake-case",
		func(t *testing.T) {
			// copied from https://github.com/blakeembrey/change-case/blob/040a079f007879cb0472ba4f7cc2e1d3185e90ba/packages/snake-case/src/index.spec.ts
			// as used in anchor.
			testCases := [][2]string{
				{"", ""},
				{"_id", "id"},
				{"test", "test"},
				{"test string", "test_string"},
				{"Test String", "test_string"},
				{"TestV2", "test_v2"},
				{"version 1.2.10", "version_1_2_10"},
				{"version 1.21.0", "version_1_21_0"},
				{"doSomething2", "do_something2"},
			}

			for _, testCase := range testCases {
				t.Run(
					testCase[0],
					func(t *testing.T) {
						assert.Equal(t,
							testCase[1],
							ToSnakeForSighash(testCase[0]),
							"from %q", testCase[0],
						)
					})
			}
		},
	)
	t.Run(
		"rust",
		// "rust package: https://docs.rs/heck",
		func(t *testing.T) {
			// copied from https://github.com/withoutboats/heck/blob/dbcfc7b8db8e532d1fad44518cf73e88d5212161/src/snake.rs#L60
			// as used in anchor.
			testCases := [][2]string{
				{"CamelCase", "camel_case"},
				{"This is Human case.", "this_is_human_case"},
				{"MixedUP CamelCase, with some Spaces", "mixed_up_camel_case_with_some_spaces"},
				{"mixed_up_ snake_case with some _spaces", "mixed_up_snake_case_with_some_spaces"},
				{"kebab-case", "kebab_case"},
				{"SHOUTY_SNAKE_CASE", "shouty_snake_case"},
				{"snake_case", "snake_case"},
				{"this-contains_ ALLKinds OfWord_Boundaries", "this_contains_all_kinds_of_word_boundaries"},

				// #[cfg(feature = "unicode")]
				{"XΣXΣ baﬄe", "xσxσ_baﬄe"},
				{"XMLHttpRequest", "xml_http_request"},
				{"FIELD_NAME11", "field_name11"},
				{"99BOTTLES", "99bottles"},
				{"FieldNamE11", "field_nam_e11"},

				{"abc123def456", "abc123def456"},
				{"abc123DEF456", "abc123_def456"},
				{"abc123Def456", "abc123_def456"},
				{"abc123DEf456", "abc123_d_ef456"},
				{"ABC123def456", "abc123def456"},
				{"ABC123DEF456", "abc123def456"},
				{"ABC123Def456", "abc123_def456"},
				{"ABC123DEf456", "abc123d_ef456"},
				{"ABC123dEEf456FOO", "abc123d_e_ef456_foo"},
				{"abcDEF", "abc_def"},
				{"ABcDE", "a_bc_de"},
			}

			for _, testCase := range testCases {
				t.Run(
					testCase[0],
					func(t *testing.T) {
						assert.Equal(t,
							testCase[1],
							ToSnakeForSighash(testCase[0]),
							"from %q", testCase[0],
						)
					})
			}
		})

}
