package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

// #[serde(tag = "kind", rename_all = "lowercase")]
//
//	pub enum IdlTypeDefTy {
//	    Struct {
//	        #[serde(skip_serializing_if = "is_default")]
//	        fields: Option<IdlDefinedFields>,
//	    },
//	    Enum {
//	        variants: Vec<IdlEnumVariant>,
//	    },
//	    Type {
//	        alias: IdlType,
//	    },
//	}
type IdlTypeDefTy interface {
	_is_IdlTypeDefTy()
}

// export type IdlTypeDefTy =
//   | IdlTypeDefTyEnum
//   | IdlTypeDefTyStruct
//   | IdlTypeDefTyType;

// export type IdlTypeDefTyStruct = {
//   kind: "struct";
//   fields?: IdlDefinedFields;
// };

// export type IdlTypeDefTyEnum = {
//   kind: "enum";
//   variants: IdlEnumVariant[];
// };

// export type IdlTypeDefTyType = {
//   kind: "type";
//   alias: IdlType;
// };

type IdlTypeDefTyStruct struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub fields: Option<IdlDefinedFields>,
	Fields IdlDefinedFields `json:"fields,omitzero"`
}

func (IdlTypeDefTyStruct) _is_IdlTypeDefTy() {}

func (st IdlTypeDefTyStruct) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind   string           `json:"kind"`
		Fields IdlDefinedFields `json:"fields,omitzero"`
	}{
		Kind:   "struct",
		Fields: st.Fields,
	})
}

func (st *IdlTypeDefTyStruct) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"kind",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Kind   string          `json:"kind"`
		Fields json.RawMessage `json:"fields,omitzero"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "struct" {
		return fmt.Errorf("expected kind 'struct', got %s", alias.Kind)
	}
	st.Kind = "struct"
	if len(alias.Fields) > 0 {
		var fields IdlDefinedFields
		err = into_IdlDefinedFields(&fields, alias.Fields)
		if err != nil {
			return err
		}
		st.Fields = fields
	}
	return nil
}

type IdlTypeDefTyEnum struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    pub variants: Vec<IdlEnumVariant>,
	Variants VariantSlice `json:"variants"`
}

type VariantSlice []IdlEnumVariant

func (obj IdlTypeDefTyEnum) IsAllSimple() bool {
	return obj.Variants.IsAllSimple()
}

func (sl VariantSlice) IsAllSimple() bool {
	for _, v := range sl {
		if !v.IsSimple() {
			return false
		}
	}
	return true
}

func (IdlTypeDefTyEnum) _is_IdlTypeDefTy() {}
func (et IdlTypeDefTyEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind     string       `json:"kind"`
		Variants VariantSlice `json:"variants"`
	}{
		Kind:     "enum",
		Variants: et.Variants,
	})
}

func (et *IdlTypeDefTyEnum) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"kind",
		"variants",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Kind     string           `json:"kind"`
		Variants []IdlEnumVariant `json:"variants"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "enum" {
		return fmt.Errorf("expected kind 'enum', got %s", alias.Kind)
	}
	et.Kind = "enum"
	et.Variants = alias.Variants
	return nil
}

func into_IdlTypeDefTy(
	dst *IdlTypeDefTy,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlTypeDefTy[*IdlTypeDefTyStruct],
		tryUnmarshal_IdlTypeDefTy[*IdlTypeDefTyEnum],
	)
}

func tryUnmarshal_IdlTypeDefTy[T IdlTypeDefTy](
	data []byte,
) (IdlTypeDefTy, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
