package idltype

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

//	pub enum IdlGenericArg {
//	    Type {
//	        #[serde(rename = "type")]
//	        ty: IdlType,
//	    },
//	    Const {
//	        value: String,
//	    },
//	}
type IdlGenericArg interface {
	_is_IdlGenericArg()
}

// export type IdlGenericArg = IdlGenericArgType | IdlGenericArgConst;

// export type IdlGenericArgType = { kind: "type"; type: IdlType };

// export type IdlGenericArgConst = { kind: "const"; value: string };

type IdlGenericArgType struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    #[serde(rename = "type")]
	//	    pub ty: IdlType,
	Ty IdlType `json:"type"`
}

func (IdlGenericArgType) _is_IdlGenericArg() {}

func (i IdlGenericArgType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string  `json:"kind"`
		Ty   IdlType `json:"type"`
	}{
		Kind: "type",
		Ty:   i.Ty,
	})
}

func (i *IdlGenericArgType) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"type",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		//	    pub kind: String,
		Kind string          `json:"kind"`
		Ty   json.RawMessage `json:"type"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "type" {
		return fmt.Errorf("expected kind 'type', got %s", alias.Kind)
	}
	i.Kind = alias.Kind
	{
		err = Into(&i.Ty, alias.Ty)
		if err != nil {
			return err
		}
	}
	return nil
}

type IdlGenericArgConst struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    pub value: String,
	Value string `json:"value"`
}

func (IdlGenericArgConst) _is_IdlGenericArg() {}
func (i IdlGenericArgConst) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind  string `json:"kind"`
		Value string `json:"value"`
	}{
		Kind:  "const",
		Value: i.Value,
	})
}

func (i *IdlGenericArgConst) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"value",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		//	    pub kind: String,
		Kind  string `json:"kind"`
		Value string `json:"value"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "const" {
		return fmt.Errorf("expected kind 'const', got %s", alias.Kind)
	}
	i.Kind = alias.Kind
	i.Value = alias.Value
	return nil
}

func Into_IdlGenericArg(
	dst *IdlGenericArg,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlGenericArg[*IdlGenericArgType],
		tryUnmarshal_IdlGenericArg[*IdlGenericArgConst],
	)
}

func tryUnmarshal_IdlGenericArg[T IdlGenericArg](
	data []byte,
) (IdlGenericArg, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
