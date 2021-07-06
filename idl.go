package main

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	. "github.com/gagliardetto/utilz"
)

// https://github.com/project-serum/anchor/blob/97e9e03fb041b8b888a9876a7c0676d9bb4736f3/ts/src/idl.ts
type IDL struct {
	Version      string           `json:"version"`
	Name         string           `json:"name"`
	Instructions []IdlInstruction `json:"instructions"`
	State        *IdlState        `json:"state,omitempty"`
	Accounts     []IdlTypeDef     `json:"accounts,omitempty"`
	Types        []IdlTypeDef     `json:"types,omitempty"`
	Events       []IdlEvent       `json:"events,omitempty"`
	Errors       []IdlErrorCode   `json:"errors,omitempty"`
}

type IdlEvent struct {
	Name   string          `json:"name"`
	Fields []IdlEventField `json:"fields"`
}

type IdlEventField struct {
	Name  string          `json:"name"`
	Type  IdlTypeEnvelope `json:"type"`
	Index bool            `json:"index"`
}

type IdlInstruction struct {
	Name     string           `json:"name"`
	Accounts []IdlAccountItem `json:"accounts"`
	Args     []IdlField       `json:"args"`
}

type IdlState struct {
	Struct  IdlTypeDef       `json:"struct"`
	Methods []IdlStateMethod `json:"methods"`
}

type IdlStateMethod = IdlInstruction

// type IdlAccountItem = IdlAccount | IdlAccounts;
type IdlAccountItem struct {
	IdlAccount  *IdlAccount
	IdlAccounts *IdlAccounts
}

// TODO: verify with examples
func (env *IdlAccountItem) UnmarshalJSON(data []byte) error {

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		return fmt.Errorf("envelope is nil: %v", env)
	}

	switch v := temp.(type) {
	case map[string]interface{}:
		{
			Ln(ShakespeareBG("::IdlAccountItem"))
			spew.Dump(v)

			if len(v) == 0 {
				return nil
			}

			if _, ok := v["accounts"]; ok {
				if err := TranscodeJSON(temp, &env.IdlAccounts); err != nil {
					return err
				}
			}
			// TODO: check both isMut and isSigner
			if _, ok := v["isMut"]; ok {
				if err := TranscodeJSON(temp, &env.IdlAccount); err != nil {
					return err
				}
			}

			// panic(Sf("what is this?:\n%s", spew.Sdump(temp)))
		}
	default:
		return fmt.Errorf("Unknown kind: %s", spew.Sdump(temp))
	}

	return nil
}

type IdlAccount struct {
	Name     string `json:"name"`
	IsMut    bool   `json:"isMut"`
	IsSigner bool   `json:"isSigner"`
}

// A nested/recursive version of IdlAccount.
type IdlAccounts struct {
	Name     string           `json:"name"`
	Accounts []IdlAccountItem `json:"accounts"`
}

type IdlField struct {
	Name string          `json:"name"`
	Type IdlTypeEnvelope `json:"type"`
}

type IdlTypeDef struct {
	Name string       `json:"name"`
	Type IdlTypeDefTy `json:"type"`
}

type IdlTypeDefTyKind string

const (
	IdlTypeDefTyKindStruct IdlTypeDefTyKind = "struct"
	IdlTypeDefTyKindEnum   IdlTypeDefTyKind = "enum"
)

type IdlTypeDefTy struct {
	Kind     IdlTypeDefTyKind  `json:"kind"`
	Fields   *IdlTypeDefStruct `json:"fields,omitempty"`
	Variants []IdlEnumVariant  `json:"variants,omitempty"`
}

type IdlTypeDefStruct = []IdlField

type IdlTypeAsString string

const (
	IdlTypeBool      IdlTypeAsString = "bool"
	IdlTypeU8        IdlTypeAsString = "u8"
	IdlTypeI8        IdlTypeAsString = "i8"
	IdlTypeU16       IdlTypeAsString = "u16"
	IdlTypeI16       IdlTypeAsString = "i16"
	IdlTypeU32       IdlTypeAsString = "u32"
	IdlTypeI32       IdlTypeAsString = "i32"
	IdlTypeU64       IdlTypeAsString = "u64"
	IdlTypeI64       IdlTypeAsString = "i64"
	IdlTypeU128      IdlTypeAsString = "u128"
	IdlTypeI128      IdlTypeAsString = "i128"
	IdlTypeBytes     IdlTypeAsString = "bytes"
	IdlTypeString    IdlTypeAsString = "string"
	IdlTypePublicKey IdlTypeAsString = "publicKey"
	// | IdlTypeVec
	// | IdlTypeOption
	// | IdlTypeDefined;
)

type IdlTypeVec struct {
	Vec IdlTypeEnvelope `json:"vec"`
}

type IdlTypeOption struct {
	Option IdlTypeEnvelope `json:"option"`
}

// User defined type.
type IdlTypeDefined struct {
	Defined string `json:"defined"`
}

// Wrapper type:
type IdlTypeEnvelopeArray struct {
	Thing IdlTypeEnvelope
	Num   float64
}

func (env *IdlTypeEnvelope) UnmarshalJSON(data []byte) error {

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		return fmt.Errorf("envelope is nil: %v", env)
	}

	switch v := temp.(type) {
	case string:
		{
			env.asString = IdlTypeAsString(v)
		}
	case map[string]interface{}:
		{
			Ln(PurpleBG("::IdlTypeEnvelope"))
			spew.Dump(v)

			if len(v) == 0 {
				return nil
			}

			if _, ok := v["vec"]; ok {
				var target IdlTypeVec
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.asIdlTypeVec = &target
			}
			if _, ok := v["option"]; ok {
				var target IdlTypeOption
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.asIdlTypeOption = &target
			}
			if _, ok := v["defined"]; ok {
				var target IdlTypeDefined
				if err := TranscodeJSON(temp, &target); err != nil {
					return err
				}
				env.asIdlTypeDefined = &target
			}
			if got, ok := v["array"]; ok {

				if _, ok := got.([]interface{}); !ok {
					panic(Sf("array is not in expected format:\n%s", spew.Sdump(got)))
				}
				arrVal := got.([]interface{})
				if len(arrVal) != 2 {
					panic(Sf("array is not of expected length:\n%s", spew.Sdump(got)))
				}
				var target IdlTypeEnvelopeArray
				if err := TranscodeJSON(arrVal[0], &target.Thing); err != nil {
					return err
				}

				target.Num = arrVal[1].(float64)

				env.asIdlTypeEnvelopeArray = &target
			}
			// panic(Sf("what is this?:\n%s", spew.Sdump(temp)))
		}
	default:
		return fmt.Errorf("Unknown kind: %s", spew.Sdump(temp))
	}

	return nil
}

// Wrapper type:
type IdlTypeEnvelope struct {
	asString               IdlTypeAsString
	asIdlTypeVec           *IdlTypeVec
	asIdlTypeOption        *IdlTypeOption
	asIdlTypeDefined       *IdlTypeDefined
	asIdlTypeEnvelopeArray *IdlTypeEnvelopeArray
}

func (env *IdlTypeEnvelope) IsString() bool {
	return env.asString != ""
}
func (env *IdlTypeEnvelope) IsIdlTypeVec() bool {
	return env.asIdlTypeVec != nil
}
func (env *IdlTypeEnvelope) IsIdlTypeOption() bool {
	return env.asIdlTypeOption != nil
}
func (env *IdlTypeEnvelope) IsIdlTypeDefined() bool {
	return env.asIdlTypeDefined != nil
}
func (env *IdlTypeEnvelope) IsArray() bool {
	return env.asIdlTypeEnvelopeArray != nil
}

// Getters:
func (env *IdlTypeEnvelope) GetString() IdlTypeAsString {
	return env.asString
}
func (env *IdlTypeEnvelope) GetIdlTypeVec() *IdlTypeVec {
	return env.asIdlTypeVec
}
func (env *IdlTypeEnvelope) GetIdlTypeOption() *IdlTypeOption {
	return env.asIdlTypeOption
}
func (env *IdlTypeEnvelope) GetIdlTypeDefined() *IdlTypeDefined {
	return env.asIdlTypeDefined
}
func (env *IdlTypeEnvelope) GetArray() *IdlTypeEnvelopeArray {
	return env.asIdlTypeEnvelopeArray
}

type IdlEnumVariant struct {
	Name   string         `json:"name"`
	Fields *IdlEnumFields `json:"fields,omitempty"`
}

// TODO
// type IdlEnumFields = IdlEnumFieldsNamed | IdlEnumFieldsTuple;
type IdlEnumFields struct {
	IdlEnumFieldsNamed *IdlEnumFieldsNamed
	IdlEnumFieldsTuple *IdlEnumFieldsTuple
}

// TODO: verify with examples
func (env *IdlEnumFields) UnmarshalJSON(data []byte) error {

	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp == nil {
		return fmt.Errorf("envelope is nil: %v", env)
	}

	switch v := temp.(type) {
	case []interface{}:
		{
			Ln(LimeBG("::IdlEnumFields"))
			spew.Dump(v)

			if len(v) == 0 {
				return nil
			}

			firstItem := v[0]

			if _, ok := firstItem.(map[string]interface{})["name"]; ok {
				// TODO:
				// If has `name` field, then it's most likely a IdlEnumFieldsNamed.
				if err := TranscodeJSON(temp, &env.IdlEnumFieldsNamed); err != nil {
					return err
				}
			} else {
				if err := TranscodeJSON(temp, &env.IdlEnumFieldsTuple); err != nil {
					return err
				}
			}

			// panic(Sf("what is this?:\n%s", spew.Sdump(temp)))
		}
	default:
		return fmt.Errorf("Unknown kind: %s", spew.Sdump(temp))
	}

	return nil
}

type IdlEnumFieldsNamed []IdlField

type IdlEnumFieldsTuple []IdlTypeEnvelope

type IdlErrorCode struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Msg  string `json:"msg,omitempty"`
}
