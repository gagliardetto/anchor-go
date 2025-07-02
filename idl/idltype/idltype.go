package idltype

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

//	pub enum IdlType {
//	    Bool,
//	    U8,
//	    I8,
//	    U16,
//	    I16,
//	    U32,
//	    I32,
//	    F32,
//	    U64,
//	    I64,
//	    F64,
//	    U128,
//	    I128,
//	    U256,
//	    I256,
//	    Bytes,
//	    String,
//	    Pubkey,
//	    Option(Box<IdlType>),
//	    Vec(Box<IdlType>),
//	    Array(Box<IdlType>, IdlArrayLen),
//	    Defined {
//	        name: String,
//	        #[serde(default, skip_serializing_if = "is_default")]
//	        generics: Vec<IdlGenericArg>,
//	    },
//	    Generic(String),
//	}

type IdlType interface {
	_is_IdlType()
	String() string
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// "bool" => IdlType::Bool,
// "u8" => IdlType::U8,
// "i8" => IdlType::I8,
// "u16" => IdlType::U16,
// "i16" => IdlType::I16,
// "u32" => IdlType::U32,
// "i32" => IdlType::I32,
// "f32" => IdlType::F32,
// "u64" => IdlType::U64,
// "i64" => IdlType::I64,
// "f64" => IdlType::F64,
// "u128" => IdlType::U128,
// "i128" => IdlType::I128,
// "u256" => IdlType::U256,
// "i256" => IdlType::I256,
// "Vec<u8>" => IdlType::Bytes,
// "String" | "&str" | "&'staticstr" => IdlType::String,
// "Pubkey" => IdlType::Pubkey,

type Bool struct {
	_ struct{}
}

func (Bool) _is_IdlType() {}
func (Bool) String() string {
	return "bool"
}

func (Bool) MarshalJSON() ([]byte, error) {
	return []byte(`"bool"`), nil
}

func (Bool) UnmarshalJSON(data []byte) error {
	if string(data) != `"bool"` {
		return fmt.Errorf(`expected "bool", got %s`, string(data))
	}
	return nil
}

type U8 struct {
	_ struct{}
}

func (U8) _is_IdlType() {}
func (U8) String() string {
	return "u8"
}

func (U8) MarshalJSON() ([]byte, error) {
	return []byte(`"u8"`), nil
}

func (U8) UnmarshalJSON(data []byte) error {
	if string(data) != `"u8"` {
		return fmt.Errorf(`expected "u8", got %s`, string(data))
	}
	return nil
}

type I8 struct {
	_ struct{}
}

func (I8) _is_IdlType() {}
func (I8) String() string {
	return "i8"
}

func (I8) MarshalJSON() ([]byte, error) {
	return []byte(`"i8"`), nil
}

func (I8) UnmarshalJSON(data []byte) error {
	if string(data) != `"i8"` {
		return fmt.Errorf(`expected "i8", got %s`, string(data))
	}
	return nil
}

type U16 struct {
	_ struct{}
}

func (U16) _is_IdlType() {}
func (U16) String() string {
	return "u16"
}

func (U16) MarshalJSON() ([]byte, error) {
	return []byte(`"u16"`), nil
}

func (U16) UnmarshalJSON(data []byte) error {
	if string(data) != `"u16"` {
		return fmt.Errorf(`expected "u16", got %s`, string(data))
	}
	return nil
}

type I16 struct {
	_ struct{}
}

func (I16) _is_IdlType() {}
func (I16) String() string {
	return "i16"
}

func (I16) MarshalJSON() ([]byte, error) {
	return []byte(`"i16"`), nil
}

func (I16) UnmarshalJSON(data []byte) error {
	if string(data) != `"i16"` {
		return fmt.Errorf(`expected "i16", got %s`, string(data))
	}
	return nil
}

type U32 struct {
	_ struct{}
}

func (U32) _is_IdlType() {}
func (U32) String() string {
	return "u32"
}

func (U32) MarshalJSON() ([]byte, error) {
	return []byte(`"u32"`), nil
}

func (U32) UnmarshalJSON(data []byte) error {
	if string(data) != `"u32"` {
		return fmt.Errorf(`expected "u32", got %s`, string(data))
	}
	return nil
}

type I32 struct {
	_ struct{}
}

func (I32) _is_IdlType() {}
func (I32) String() string {
	return "i32"
}

func (I32) MarshalJSON() ([]byte, error) {
	return []byte(`"i32"`), nil
}

func (I32) UnmarshalJSON(data []byte) error {
	if string(data) != `"i32"` {
		return fmt.Errorf(`expected "i32", got %s`, string(data))
	}
	return nil
}

type F32 struct {
	_ struct{}
}

func (F32) _is_IdlType() {}
func (F32) String() string {
	return "f32"
}

func (F32) MarshalJSON() ([]byte, error) {
	return []byte(`"f32"`), nil
}

func (F32) UnmarshalJSON(data []byte) error {
	if string(data) != `"f32"` {
		return fmt.Errorf(`expected "f32", got %s`, string(data))
	}
	return nil
}

type U64 struct {
	_ struct{}
}

func (U64) _is_IdlType() {}
func (U64) String() string {
	return "u64"
}

func (U64) MarshalJSON() ([]byte, error) {
	return []byte(`"u64"`), nil
}

func (U64) UnmarshalJSON(data []byte) error {
	if string(data) != `"u64"` {
		return fmt.Errorf(`expected "u64", got %s`, string(data))
	}
	return nil
}

type I64 struct {
	_ struct{}
}

func (I64) _is_IdlType() {}
func (I64) String() string {
	return "i64"
}

func (I64) MarshalJSON() ([]byte, error) {
	return []byte(`"i64"`), nil
}

func (I64) UnmarshalJSON(data []byte) error {
	if string(data) != `"i64"` {
		return fmt.Errorf(`expected "i64", got %s`, string(data))
	}
	return nil
}

type F64 struct {
	_ struct{}
}

func (F64) _is_IdlType() {}
func (F64) String() string {
	return "f64"
}

func (F64) MarshalJSON() ([]byte, error) {
	return []byte(`"f64"`), nil
}

func (F64) UnmarshalJSON(data []byte) error {
	if string(data) != `"f64"` {
		return fmt.Errorf(`expected "f64", got %s`, string(data))
	}
	return nil
}

type U128 struct {
	_ struct{}
}

func (U128) _is_IdlType() {}
func (U128) String() string {
	return "u128"
}

func (U128) MarshalJSON() ([]byte, error) {
	return []byte(`"u128"`), nil
}

func (U128) UnmarshalJSON(data []byte) error {
	if string(data) != `"u128"` {
		return fmt.Errorf(`expected "u128", got %s`, string(data))
	}
	return nil
}

type I128 struct {
	_ struct{}
}

func (I128) _is_IdlType() {}
func (I128) String() string {
	return "i128"
}

func (I128) MarshalJSON() ([]byte, error) {
	return []byte(`"i128"`), nil
}

func (I128) UnmarshalJSON(data []byte) error {
	if string(data) != `"i128"` {
		return fmt.Errorf(`expected "i128", got %s`, string(data))
	}
	return nil
}

type U256 struct {
	_ struct{}
}

func (U256) _is_IdlType() {}
func (U256) String() string {
	return "u256"
}

func (U256) MarshalJSON() ([]byte, error) {
	return []byte(`"u256"`), nil
}

func (U256) UnmarshalJSON(data []byte) error {
	if string(data) != `"u256"` {
		return fmt.Errorf(`expected "u256", got %s`, string(data))
	}
	return nil
}

type I256 struct {
	_ struct{}
}

func (I256) _is_IdlType() {}
func (I256) String() string {
	return "i256"
}

func (I256) MarshalJSON() ([]byte, error) {
	return []byte(`"i256"`), nil
}

func (I256) UnmarshalJSON(data []byte) error {
	if string(data) != `"i256"` {
		return fmt.Errorf(`expected "i256", got %s`, string(data))
	}
	return nil
}

type Bytes struct {
	_ struct{}
}

func (Bytes) _is_IdlType() {}
func (Bytes) String() string {
	return "bytes"
}

func (Bytes) MarshalJSON() ([]byte, error) {
	return []byte(`"bytes"`), nil
}

func (Bytes) UnmarshalJSON(data []byte) error {
	if string(data) != `"bytes"` {
		return fmt.Errorf(`expected "bytes", got %s`, string(data))
	}
	return nil
}

type String struct {
	_ struct{}
}

func (String) _is_IdlType() {}
func (String) String() string {
	return "string"
}

func (String) MarshalJSON() ([]byte, error) {
	return []byte(`"string"`), nil
}

func (String) UnmarshalJSON(data []byte) error {
	if string(data) != `"string"` {
		return fmt.Errorf(`expected "string", got %s`, string(data))
	}
	return nil
}

type Pubkey struct {
	_ struct{}
}

func (Pubkey) _is_IdlType() {}
func (Pubkey) String() string {
	return "pubkey"
}

func (Pubkey) MarshalJSON() ([]byte, error) {
	return []byte(`"pubkey"`), nil
}

func (Pubkey) UnmarshalJSON(data []byte) error {
	if string(data) != `"pubkey"` {
		return fmt.Errorf(`expected "pubkey", got %s`, string(data))
	}
	return nil
}

type Option struct {
	_      struct{}
	Option IdlType
}

func (o Option) _is_IdlType() {}
func (o Option) String() string {
	return fmt.Sprintf("Option<%s>", o.Option.String())
}

func (o Option) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Option IdlType `json:"option"`
		}{
			Option: o.Option,
		},
	)
}

func (o *Option) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.Option = nil
		return nil
	}
	type Alias struct {
		Option json.RawMessage `json:"option"`
	}
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if err := Into(
		&o.Option,
		tmp.Option,
	); err != nil {
		return err
	}
	return nil
}

type COption struct {
	_       struct{}
	COption IdlType
}

func (o COption) _is_IdlType() {}
func (o COption) String() string {
	return fmt.Sprintf("COption<%s>", o.COption.String())
}

func (o COption) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			COption IdlType `json:"coption"`
		}{
			COption: o.COption,
		},
	)
}

func (o *COption) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.COption = nil
		return nil
	}
	type Alias struct {
		COption json.RawMessage `json:"coption"`
	}
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if err := Into(
		&o.COption,
		tmp.COption,
	); err != nil {
		return err
	}
	return nil
}

type Vec struct {
	_   struct{}
	Vec IdlType
}

func (v Vec) _is_IdlType() {}
func (v Vec) String() string {
	return fmt.Sprintf("Vec<%s>", v.Vec.String())
}

func (v Vec) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Vec IdlType `json:"vec"`
		}{
			Vec: v.Vec,
		},
	)
}

func (v *Vec) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		v.Vec = nil
		return nil
	}
	type Alias struct {
		Vec json.RawMessage `json:"vec"`
	}
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if err := Into(
		&v.Vec,
		tmp.Vec,
	); err != nil {
		return err
	}
	return nil
}

type Array struct {
	_    struct{}
	Type IdlType
	Size IdlArrayLen
}

func (a Array) _is_IdlType() {}
func (a Array) String() string {
	return fmt.Sprintf("Array<%s, %s>", a.Type.String(), a.Size.String())
}

func (a Array) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Array []any `json:"array"`
		}{
			Array: []any{
				a.Type,
				a.Size,
			},
		},
	)
}

func (a *Array) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		a.Type = nil
		return nil
	}
	type Alias struct {
		Array []json.RawMessage `json:"array"`
	}
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if len(tmp.Array) != 2 {
		return fmt.Errorf("expected 2 elements, got %d", len(tmp.Array))
	}
	if err := Into(
		&a.Type,
		tmp.Array[0],
	); err != nil {
		return err
	}
	if err := IntoArrayLen(
		&a.Size,
		tmp.Array[1],
	); err != nil {
		return err
	}
	return nil
}

type IdlArrayLen interface {
	_is_IdlArrayLen()
	String() string
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

func IntoArrayLen(
	dst *IdlArrayLen,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlArrayLen[*IdlArrayLenGeneric],
		tryUnmarshal_IdlArrayLen[*IdlArrayLenValue],
	)
}

func tryUnmarshal_IdlArrayLen[T IdlArrayLen](data []byte) (IdlArrayLen, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}

// export type IdlArrayLen = IdlArrayLenGeneric | IdlArrayLenValue;
// export type IdlArrayLenGeneric = {
//   generic: string;
// };
// export type IdlArrayLenValue = number;

type IdlArrayLenGeneric struct {
	_       struct{}
	Generic string
}

func (IdlArrayLenGeneric) _is_IdlArrayLen() {}
func (a IdlArrayLenGeneric) String() string {
	return fmt.Sprintf("IdlArrayLenGeneric<%s>", a.Generic)
}

func (a IdlArrayLenGeneric) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Generic string `json:"generic"`
		}{
			Generic: a.Generic,
		},
	)
}

func (a *IdlArrayLenGeneric) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"generic",
	)
	if err != nil {
		return err
	}
	var tmp struct {
		Generic string `json:"generic"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp.Generic == "" {
		return fmt.Errorf("expected generic, got empty string")
	}
	a.Generic = tmp.Generic
	return nil
}

type IdlArrayLenValue struct {
	_     struct{}
	Value int
}

func (IdlArrayLenValue) _is_IdlArrayLen() {}
func (a IdlArrayLenValue) String() string {
	return fmt.Sprintf("IdlArrayLenValue<%d>", a.Value)
}

func (a IdlArrayLenValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Value)
}

func (a *IdlArrayLenValue) UnmarshalJSON(data []byte) error {
	var tmp int
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp < 0 {
		return fmt.Errorf("expected positive integer, got %d", tmp)
	}
	a.Value = tmp
	return nil
}

// export type IdlTypeDefined = {
//   defined: {
//     name: string;
//     generics?: IdlGenericArg[];
//   };
// };

type Defined struct {
	_        struct{}
	Name     string
	Generics []IdlGenericArg
}

func (Defined) _is_IdlType() {}
func (a Defined) String() string {
	return fmt.Sprintf("Defined<%s>", a.Name)
}

func (a Defined) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Defined struct {
				Name     string          `json:"name"`
				Generics []IdlGenericArg `json:"generics,omitzero"`
			} `json:"defined"`
		}{
			Defined: struct {
				Name     string          `json:"name"`
				Generics []IdlGenericArg `json:"generics,omitzero"`
			}{
				Name:     a.Name,
				Generics: a.Generics,
			},
		},
	)
}

func (a *Defined) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		a.Name = ""
		return nil
	}
	var tmp struct {
		Defined struct {
			Name     string            `json:"name"`
			Generics []json.RawMessage `json:"generics,omitzero"`
		} `json:"defined"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp.Defined.Name == "" {
		return fmt.Errorf("expected name, got empty string")
	}
	a.Name = tmp.Defined.Name
	if len(tmp.Defined.Generics) > 0 {
		a.Generics = make([]IdlGenericArg, len(tmp.Defined.Generics))
		for i, raw := range tmp.Defined.Generics {
			err := Into_IdlGenericArg(
				&a.Generics[i],
				raw,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// export type IdlTypeGeneric = {
//   generic: string;
// };

type Generic struct {
	_       struct{}
	Generic string
}

func (Generic) _is_IdlType() {}
func (a Generic) String() string {
	return fmt.Sprintf("IdlTypeGeneric<%s>", a.Generic)
}

func (a Generic) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"generic": %q}`, a.Generic)), nil
}

func (a *Generic) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"generic",
	)
	if err != nil {
		return err
	}
	var tmp struct {
		Generic string `json:"generic"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp.Generic == "" {
		return fmt.Errorf("expected generic, got empty string")
	}
	a.Generic = tmp.Generic
	return nil
}
