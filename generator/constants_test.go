package generator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenConstants(t *testing.T) {
	tests := []struct {
		name        string
		constants   []idl.IdlConst
		expectError bool
		expectCode  []string // 期望在生成的代码中找到的字符串
	}{
		{
			name: "String constant",
			constants: []idl.IdlConst{
				{
					Name:  "TEST_STRING",
					Ty:    &idltype.String{},
					Value: `"hello world"`,
				},
			},
			expectCode: []string{
				"const TEST_STRING = \"hello world\"",
			},
		},
		{
			name: "Boolean constants",
			constants: []idl.IdlConst{
				{
					Name:  "IS_ENABLED",
					Ty:    &idltype.Bool{},
					Value: "true",
				},
				{
					Name:  "IS_DISABLED",
					Ty:    &idltype.Bool{},
					Value: "false",
				},
			},
			expectCode: []string{
				"var IS_ENABLED = true",
				"var IS_DISABLED = false",
			},
		},
		{
			name: "Unsigned integer constants",
			constants: []idl.IdlConst{
				{
					Name:  "MAX_U8",
					Ty:    &idltype.U8{},
					Value: "255",
				},
				{
					Name:  "MAX_U16",
					Ty:    &idltype.U16{},
					Value: "65535",
				},
				{
					Name:  "MAX_U32",
					Ty:    &idltype.U32{},
					Value: "4294967295",
				},
				{
					Name:  "MAX_U64",
					Ty:    &idltype.U64{},
					Value: "18446744073709551615",
				},
			},
			expectCode: []string{
				"const MAX_U8 = uint8(0xff)",
				"const MAX_U16 = uint16(0xffff)",
				"const MAX_U32 = uint32(0xffffffff)",
				"const MAX_U64 = uint64(0xffffffffffffffff)",
			},
		},
		{
			name: "Signed integer constants",
			constants: []idl.IdlConst{
				{
					Name:  "MIN_I8",
					Ty:    &idltype.I8{},
					Value: "-128",
				},
				{
					Name:  "MIN_I16",
					Ty:    &idltype.I16{},
					Value: "-32768",
				},
				{
					Name:  "MIN_I32",
					Ty:    &idltype.I32{},
					Value: "-2147483648",
				},
				{
					Name:  "MIN_I64",
					Ty:    &idltype.I64{},
					Value: "-9223372036854775808",
				},
			},
			expectCode: []string{
				"const MIN_I8 = int8(-128)",
				"const MIN_I16 = int16(-32768)",
				"const MIN_I32 = int32(-2147483648)",
				"const MIN_I64 = int64(-9223372036854775808)",
			},
		},
		{
			name: "Float constants",
			constants: []idl.IdlConst{
				{
					Name:  "PI_F32",
					Ty:    &idltype.F32{},
					Value: "3.14159",
				},
				{
					Name:  "E_F64",
					Ty:    &idltype.F64{},
					Value: "2.718281828459045",
				},
			},
			expectCode: []string{
				"const PI_F32 = float32(3.14159)",
				"const E_F64 = 2.718281828459045",
			},
		},
		{
			name: "Numbers with underscores",
			constants: []idl.IdlConst{
				{
					Name:  "LARGE_NUMBER",
					Ty:    &idltype.U64{},
					Value: "100_000_000",
				},
				{
					Name:  "NEGATIVE_NUMBER",
					Ty:    &idltype.I32{},
					Value: "-1_000_000",
				},
			},
			expectCode: []string{
				"const LARGE_NUMBER = uint64(0x5f5e100)",
				"const NEGATIVE_NUMBER = int32(-1000000)",
			},
		},
		{
			name: "usize constant",
			constants: []idl.IdlConst{
				{
					Name: "MAX_BIN_PER_ARRAY",
					Ty: &idltype.Defined{
						Name: "usize",
					},
					Value: "70",
				},
			},
			expectCode: []string{
				"const MAX_BIN_PER_ARRAY = uint64(0x46)",
			},
		},
		{
			name: "isize constant",
			constants: []idl.IdlConst{
				{
					Name: "MIN_BIN_ID",
					Ty: &idltype.Defined{
						Name: "isize",
					},
					Value: "-443636",
				},
			},
			expectCode: []string{
				"const MIN_BIN_ID = int64(-443636)",
			},
		},
		{
			name: "u128 constant",
			constants: []idl.IdlConst{
				{
					Name:  "MAX_BASE_FEE",
					Ty:    &idltype.U128{},
					Value: "100_000_000",
				},
			},
			expectCode: []string{
				"var MAX_BASE_FEE = func() *big.Int",
				".SetString(\"100000000\", 10)",
			},
		},
		{
			name: "i128 constant",
			constants: []idl.IdlConst{
				{
					Name:  "MIN_BALANCE",
					Ty:    &idltype.I128{},
					Value: "-1_000_000_000_000",
				},
			},
			expectCode: []string{
				"var MIN_BALANCE = func() *big.Int",
				".SetString(\"-1000000000000\", 10)",
			},
		},
		{
			name: "Bytes constant",
			constants: []idl.IdlConst{
				{
					Name:  "SEED_BYTES",
					Ty:    &idltype.Bytes{},
					Value: "[102, 101, 101, 95, 118, 97, 117, 108, 116]",
				},
			},
			expectCode: []string{
				"var SEED_BYTES = []byte{102, 101, 101, 95, 118, 97, 117, 108, 116}",
			},
		},
		{
			name: "Pubkey constant",
			constants: []idl.IdlConst{
				{
					Name:  "PROGRAM_ID",
					Ty:    &idltype.Pubkey{},
					Value: "11111111111111111111111111111112", // System Program ID
				},
			},
			expectCode: []string{
				"var PROGRAM_ID = solanago.MustPublicKeyFromBase58(\"11111111111111111111111111111112\")",
			},
		},
		{
			name: "Empty name - should be skipped",
			constants: []idl.IdlConst{
				{
					Name:  "",
					Ty:    &idltype.U8{},
					Value: "42",
				},
				{
					Name:  "VALID_CONST",
					Ty:    &idltype.U8{},
					Value: "42",
				},
			},
			expectCode: []string{
				"const VALID_CONST = uint8(0x2a)",
			},
		},
		{
			name: "Empty value - should be skipped",
			constants: []idl.IdlConst{
				{
					Name:  "EMPTY_VALUE",
					Ty:    &idltype.U8{},
					Value: "",
				},
				{
					Name:  "VALID_CONST",
					Ty:    &idltype.U8{},
					Value: "42",
				},
			},
			expectCode: []string{
				"const VALID_CONST = uint8(0x2a)",
			},
		},
		{
			name: "Unsupported defined type",
			constants: []idl.IdlConst{
				{
					Name: "CUSTOM_TYPE",
					Ty: &idltype.Defined{
						Name: "CustomType",
					},
					Value: "42",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid string format",
			constants: []idl.IdlConst{
				{
					Name:  "INVALID_STRING",
					Ty:    &idltype.String{},
					Value: "invalid string format", // 应该有引号
				},
			},
			expectError: true,
		},
		{
			name: "Invalid number format",
			constants: []idl.IdlConst{
				{
					Name:  "INVALID_NUMBER",
					Ty:    &idltype.U8{},
					Value: "not_a_number",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid u128 format",
			constants: []idl.IdlConst{
				{
					Name:  "INVALID_U128",
					Ty:    &idltype.U128{},
					Value: "not_a_number",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个最小的 IDL 结构
			idlData := &idl.Idl{
				Constants: tt.constants,
			}

			// 创建生成器
			gen := &Generator{
				idl: idlData,
				options: &GeneratorOptions{
					Package: "test",
				},
			}

			// 生成常量
			outputFile, err := gen.gen_constants()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, outputFile)

			// 获取生成的代码
			generatedCode := outputFile.File.GoString()

			// 检查期望的代码片段是否存在
			for _, expectedCode := range tt.expectCode {
				assert.Contains(t, generatedCode, expectedCode,
					"Expected code snippet not found: %s\nGenerated code:\n%s",
					expectedCode, generatedCode)
			}

			// 基本的结构检查
			assert.Contains(t, generatedCode, "package test")
			assert.Contains(t, generatedCode, "Code generated by https://github.com/gagliardetto/anchor-go")
			assert.Contains(t, generatedCode, "This file contains constants")
		})
	}
}

func TestGenConstantsWithArrays(t *testing.T) {
	// 测试数组常量
	constants := []idl.IdlConst{
		{
			Name: "BYTE_ARRAY",
			Ty: &idltype.Array{
				Type: &idltype.U8{},
				Size: &idltype.IdlArrayLenValue{Value: 3},
			},
			Value: "[1, 2, 3]",
		},
	}

	idlData := &idl.Idl{
		Constants: constants,
	}

	gen := &Generator{
		idl: idlData,
		options: &GeneratorOptions{
			Package: "test",
		},
	}

	outputFile, err := gen.gen_constants()
	require.NoError(t, err)

	generatedCode := outputFile.File.GoString()
	assert.Contains(t, generatedCode, "var BYTE_ARRAY = [3]byte{uint8(0x1), uint8(0x2), uint8(0x3)}")
}

func TestGenConstantsEdgeCases(t *testing.T) {
	t.Run("No constants", func(t *testing.T) {
		idlData := &idl.Idl{
			Constants: []idl.IdlConst{},
		}

		gen := &Generator{
			idl: idlData,
			options: &GeneratorOptions{
				Package: "test",
			},
		}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "package test")
		// 不应该包含 "Constants defined in the IDL:" 注释
		assert.NotContains(t, generatedCode, "Constants defined in the IDL:")
	})

	t.Run("Underscore cleaning", func(t *testing.T) {
		// 测试下划线清理功能
		testCases := []struct {
			value    string
			expected string
		}{
			{"1_000", "1000"},
			{"1_000_000", "1000000"},
			{"1_2_3_4", "1234"},
			{"100", "100"}, // 没有下划线
		}

		for _, tc := range testCases {
			constants := []idl.IdlConst{
				{
					Name:  "TEST_VALUE",
					Ty:    &idltype.U64{},
					Value: tc.value,
				},
			}

			idlData := &idl.Idl{
				Constants: constants,
			}

			gen := &Generator{
				idl: idlData,
				options: &GeneratorOptions{
					Package: "test",
				},
			}

			outputFile, err := gen.gen_constants()
			require.NoError(t, err, "Failed for value: %s", tc.value)

			generatedCode := outputFile.File.GoString()

			// 验证生成的代码不包含原始的下划线值
			if strings.Contains(tc.value, "_") {
				assert.NotContains(t, generatedCode, tc.value)
			}
		}
	})
}

func TestGenConstantsPerformance(t *testing.T) {
	// 测试大量常量的性能
	constants := make([]idl.IdlConst, 1000)
	for i := 0; i < 1000; i++ {
		constants[i] = idl.IdlConst{
			Name:  fmt.Sprintf("CONST_%d", i),
			Ty:    &idltype.U32{},
			Value: fmt.Sprintf("%d", i),
		}
	}

	idlData := &idl.Idl{
		Constants: constants,
	}

	gen := &Generator{
		idl: idlData,
		options: &GeneratorOptions{
			Package: "test",
		},
	}

	// 测试性能（应该在合理时间内完成）
	outputFile, err := gen.gen_constants()
	require.NoError(t, err)
	require.NotNil(t, outputFile)

	generatedCode := outputFile.File.GoString()
	assert.Contains(t, generatedCode, "CONST_0")
	assert.Contains(t, generatedCode, "CONST_999")
}

// TestGenConstantsSpecialCases 测试特殊情况
func TestGenConstantsSpecialCases(t *testing.T) {
	t.Run("Zero values", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "ZERO_U8",
				Ty:    &idltype.U8{},
				Value: "0",
			},
			{
				Name:  "ZERO_I32",
				Ty:    &idltype.I32{},
				Value: "0",
			},
			{
				Name:  "ZERO_F64",
				Ty:    &idltype.F64{},
				Value: "0.0",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "const ZERO_U8 = uint8(0x0)")
		assert.Contains(t, generatedCode, "const ZERO_I32 = int32(0)")
		assert.Contains(t, generatedCode, "const ZERO_F64 = 0")
	})

	t.Run("Maximum values", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "MAX_U8_VALUE",
				Ty:    &idltype.U8{},
				Value: "255",
			},
			{
				Name:  "MAX_I8_VALUE",
				Ty:    &idltype.I8{},
				Value: "127",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "const MAX_U8_VALUE = uint8(0xff)")
		assert.Contains(t, generatedCode, "const MAX_I8_VALUE = int8(127)")
	})

	t.Run("Complex underscores", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "COMPLEX_NUMBER",
				Ty:    &idltype.U64{},
				Value: "1_000_000_000_000_000_000",
			},
			{
				Name:  "HEX_LIKE_NUMBER",
				Ty:    &idltype.U32{},
				Value: "0_x_F_F_F_F", // 这不是真正的十六进制，只是包含下划线的数字
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		// 第二个应该失败，因为它不是有效的数字
		outputFile, err := gen.gen_constants()
		assert.Error(t, err) // 应该失败
		_ = outputFile
	})

	t.Run("Scientific notation", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "SCIENTIFIC_F32",
				Ty:    &idltype.F32{},
				Value: "1.23e-4",
			},
			{
				Name:  "SCIENTIFIC_F64",
				Ty:    &idltype.F64{},
				Value: "1.23456789e10",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "const SCIENTIFIC_F32 = float32(0.000123)")
		assert.Contains(t, generatedCode, "const SCIENTIFIC_F64 = 1.23456789e+10")
	})

	t.Run("Empty bytes array", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "EMPTY_BYTES",
				Ty:    &idltype.Bytes{},
				Value: "[]",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "var EMPTY_BYTES = []byte{}")
	})

	t.Run("With docs", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "DOCUMENTED_CONST",
				Docs:  []string{"This is a test constant", "With multiple lines of documentation"},
				Ty:    &idltype.U32{},
				Value: "42",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "// This is a test constant")
		assert.Contains(t, generatedCode, "// With multiple lines of documentation")
		assert.Contains(t, generatedCode, "const DOCUMENTED_CONST = uint32(0x2a)")
	})
}

// TestGenConstantsErrorCases 测试各种错误情况
func TestGenConstantsErrorCases(t *testing.T) {
	t.Run("Invalid pubkey", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "INVALID_PUBKEY",
				Ty:    &idltype.Pubkey{},
				Value: "invalid_pubkey_format",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		_, err := gen.gen_constants()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse pubkey")
	})

	t.Run("Invalid bytes format", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "INVALID_BYTES",
				Ty:    &idltype.Bytes{},
				Value: "[1, 2, invalid]",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		_, err := gen.gen_constants()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal bytes")
	})

	t.Run("Invalid array format", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name: "INVALID_ARRAY",
				Ty: &idltype.Array{
					Type: &idltype.U8{},
					Size: &idltype.IdlArrayLenValue{Value: 3},
				},
				Value: "[1, 2]", // 只有2个元素，但期望3个
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		_, err := gen.gen_constants()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "got 2")
	})

	t.Run("Number overflow", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "OVERFLOW_U8",
				Ty:    &idltype.U8{},
				Value: "256", // 超出 u8 范围
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		_, err := gen.gen_constants()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse u8")
	})
}

// TestGenConstantsRealWorldExamples 测试真实世界的例子
func TestGenConstantsRealWorldExamples(t *testing.T) {
	t.Run("Solana program constants", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "LAMPORTS_PER_SOL",
				Ty:    &idltype.U64{},
				Value: "1_000_000_000",
			},
			{
				Name:  "SEED_PREFIX",
				Ty:    &idltype.String{},
				Value: `"anchor"`,
			},
			{
				Name:  "MAX_SEED_LEN",
				Ty:    &idltype.U32{},
				Value: "32",
			},
			{
				Name:  "SYSTEM_PROGRAM_ID",
				Ty:    &idltype.Pubkey{},
				Value: "11111111111111111111111111111112",
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "myprogram"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()
		assert.Contains(t, generatedCode, "package myprogram")
		assert.Contains(t, generatedCode, "const LAMPORTS_PER_SOL = uint64(0x3b9aca00)")
		assert.Contains(t, generatedCode, "const SEED_PREFIX = \"anchor\"")
		assert.Contains(t, generatedCode, "const MAX_SEED_LEN = uint32(0x20)")
		assert.Contains(t, generatedCode, "var SYSTEM_PROGRAM_ID = solanago.MustPublicKeyFromBase58")
	})

	t.Run("Mixed types with all supported features", func(t *testing.T) {
		constants := []idl.IdlConst{
			{
				Name:  "FEATURE_ENABLED",
				Docs:  []string{"Feature flag for new functionality"},
				Ty:    &idltype.Bool{},
				Value: "true",
			},
			{
				Name: "MAX_BIN_COUNT",
				Docs: []string{"Maximum number of bins per array"},
				Ty: &idltype.Defined{
					Name: "usize",
				},
				Value: "70",
			},
			{
				Name:  "PROTOCOL_FEE",
				Docs:  []string{"Protocol fee in basis points"},
				Ty:    &idltype.U128{},
				Value: "10_000_000_000_000_000_000",
			},
			{
				Name: "SIGNATURE_SEED",
				Ty: &idltype.Array{
					Type: &idltype.U8{},
					Size: &idltype.IdlArrayLenValue{Value: 8},
				},
				Value: "[115, 105, 103, 110, 97, 116, 117, 114]", // "signatur" in ASCII
			},
		}

		idlData := &idl.Idl{Constants: constants}
		gen := &Generator{idl: idlData, options: &GeneratorOptions{Package: "test"}}

		outputFile, err := gen.gen_constants()
		require.NoError(t, err)

		generatedCode := outputFile.File.GoString()

		// 检查注释
		assert.Contains(t, generatedCode, "// Feature flag for new functionality")
		assert.Contains(t, generatedCode, "// Maximum number of bins per array")
		assert.Contains(t, generatedCode, "// Protocol fee in basis points")

		// 检查生成的常量
		assert.Contains(t, generatedCode, "var FEATURE_ENABLED = true")
		assert.Contains(t, generatedCode, "const MAX_BIN_COUNT = uint64(0x46)")
		assert.Contains(t, generatedCode, "var PROTOCOL_FEE = func() *big.Int")
		assert.Contains(t, generatedCode, "var SIGNATURE_SEED = [8]byte{uint8(0x73), uint8(0x69), uint8(0x67), uint8(0x6e), uint8(0x61), uint8(0x74), uint8(0x75), uint8(0x72)}")
	})
}
