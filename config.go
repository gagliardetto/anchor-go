package main

import (
	"errors"
	"fmt"

	. "github.com/gagliardetto/utilz"
)

var conf = &Config{}

type Config struct {
	Encoding EncoderName
	TypeID   TypeIDName
	Debug    bool
	DstDir   string
	ModPath  string
}

func GetConfig() *Config {
	return conf
}

// Validate validates
func (cfg *Config) Validate() error {
	if cfg == nil {
		return errors.New("cfg is nil")
	}
	if !isValidEncoder(cfg.Encoding) {
		return fmt.Errorf("Encoder kind is not valid: %q", cfg.Encoding)
	}
	if !isValidTypeIDName(cfg.TypeID) {
		return fmt.Errorf("TypeID kind is not valid: %q", cfg.TypeID)
	}
	return nil
}

func isValidEncoder(enc EncoderName) bool {
	return SliceContains(
		[]string{
			string(EncodingBorsh),
			string(EncodingBin),
			string(EncodingCompactU16),
		},
		string(enc),
	)
}

type TypeIDName string

const (
	TypeIDUvarint32 TypeIDName = "uvarint32"
	TypeIDUint32    TypeIDName = "uint32"
	TypeIDUint8     TypeIDName = "uint8"
	TypeIDAnchor    TypeIDName = "anchor"
	TypeIDNoType    TypeIDName = "notype"
)

func isValidTypeIDName(typeID TypeIDName) bool {
	return SliceContains(
		[]string{
			string(TypeIDUvarint32),
			string(TypeIDUint32),
			string(TypeIDUint8),
			string(TypeIDAnchor),
			string(TypeIDNoType),
		},
		string(typeID),
	)
}

type TypeIDNameSlice []TypeIDName

func (slice TypeIDNameSlice) Has(v TypeIDName) bool {
	for _, enc := range slice {
		if v == enc {
			return true
		}
	}
	return false
}
func (name TypeIDName) On(
	candidates TypeIDNameSlice,
	fn func(),
) TypeIDName {
	if candidates.Has(GetConfig().TypeID) {
		fn()
	}
	return name
}

type EncoderName string

const (
	// github.com/gagliardetto/binary
	EncodingBin EncoderName = "bin"
	// github.com/gagliardetto/borsh-go
	EncodingBorsh EncoderName = "borsh"
	// https://docs.solana.com/developing/programming-model/transactions#compact-array-format
	EncodingCompactU16 EncoderName = "compact-u16"
)

func (name EncoderName) _NewEncoder() string {
	switch enc := GetConfig().Encoding; enc {
	case EncodingBin:
		return "NewBinEncoder"
	case EncodingBorsh:
		return "NewBorshEncoder"
	case EncodingCompactU16:
		return "NewCompact16Encoder"
	default:
		panic(enc)
	}
}

func (name EncoderName) _NewDecoder() string {
	switch enc := GetConfig().Encoding; enc {
	case EncodingBin:
		return "NewBinDecoder"
	case EncodingBorsh:
		return "NewBorshDecoder"
	case EncodingCompactU16:
		return "NewCompact16Decoder"
	default:
		panic(enc)
	}
}

type EncoderNameSlice []EncoderName

func (slice EncoderNameSlice) Has(v EncoderName) bool {
	for _, enc := range slice {
		if v == enc {
			return true
		}
	}
	return false
}
func (name EncoderName) On(
	anyEncoding EncoderNameSlice,
	fn func(),
) EncoderName {
	if anyEncoding.Has(GetConfig().Encoding) {
		fn()
	}
	return name
}

func (name EncoderName) OnEncodingBin(fn func()) EncoderName {
	if GetConfig().Encoding == EncodingBin {
		fn()
	}
	return name
}

func (name EncoderName) OnEncodingBorsh(fn func()) EncoderName {
	if GetConfig().Encoding == EncodingBorsh {
		fn()
	}
	return name
}

func (name EncoderName) OnEncodingCompactU16(fn func()) EncoderName {
	if GetConfig().Encoding == EncodingCompactU16 {
		fn()
	}
	return name
}
