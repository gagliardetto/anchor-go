package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

// #[serde(tag = "kind", rename_all = "lowercase")]
//
//	pub enum IdlTypeDefGeneric {
//	    Type {
//	        name: String,
//	    },
//	    Const {
//	        name: String,
//	        #[serde(rename = "type")]
//	        ty: String,
//	    },
//	}
type IdlTypeDefGeneric interface {
	_is_IdlTypeDefGeneric()
}

// export type IdlTypeDefGeneric = IdlTypeDefGenericType | IdlTypeDefGenericConst;

// export type IdlTypeDefGenericType = {
//   kind: "type";
//   name: string;
// };

// export type IdlTypeDefGenericConst = {
//   kind: "const";
//   name: string;
//   type: string;
// };

type IdlTypeDefGenericType struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    pub name: String,
	Name string `json:"name"`
}

func (IdlTypeDefGenericType) _is_IdlTypeDefGeneric() {}
func (gt IdlTypeDefGenericType) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
	}{
		Kind: "type",
		Name: gt.Name,
	})
}

func (gt *IdlTypeDefGenericType) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"kind",
		"name",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "type" {
		return fmt.Errorf("expected kind 'type', got %s", alias.Kind)
	}
	gt.Kind = "type"
	gt.Name = alias.Name
	return nil
}

type IdlTypeDefGenericConst struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    pub name: String,
	Name string `json:"name"`
	//	    #[serde(rename = "type")]
	//	    pub ty: String,
	Ty string `json:"type"`
}

func (IdlTypeDefGenericConst) _is_IdlTypeDefGeneric() {}
func (gc IdlTypeDefGenericConst) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
		Ty   string `json:"type"`
	}{
		Kind: "const",
		Name: gc.Name,
		Ty:   gc.Ty,
	})
}

func (gc *IdlTypeDefGenericConst) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"kind",
		"name",
		"type",
	)
	if err != nil {
		return err
	}

	type Alias struct {
		Kind string `json:"kind"`
		Name string `json:"name"`
		Ty   string `json:"type"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "const" {
		return fmt.Errorf("expected kind 'const', got %s", alias.Kind)
	}
	gc.Kind = "const"
	gc.Name = alias.Name
	gc.Ty = alias.Ty
	return nil
}

func into_IdlTypeDefGeneric(
	dst *IdlTypeDefGeneric,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlTypeDefGeneric[*IdlTypeDefGenericType],
		tryUnmarshal_IdlTypeDefGeneric[*IdlTypeDefGenericConst],
	)
}

func tryUnmarshal_IdlTypeDefGeneric[T IdlTypeDefGeneric](
	data []byte,
) (IdlTypeDefGeneric, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
