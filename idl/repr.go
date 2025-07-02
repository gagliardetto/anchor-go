package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

// #[serde(tag = "kind", rename_all = "lowercase")]
// #[non_exhaustive]
//
//	pub enum IdlRepr {
//	    Rust(IdlReprModifier),
//	    C(IdlReprModifier),
//	    Transparent,
//	}
type IdlRepr interface {
	_is_IdlRepr()
}

// export type IdlRepr = IdlReprRust | IdlReprC | IdlReprTransparent;

// export type IdlReprRust = {
//   kind: "rust";
// } & IdlReprModifier;

// export type IdlReprC = {
//   kind: "c";
// } & IdlReprModifier;

// export type IdlReprTransparent = {
//   kind: "transparent";
// };

// export type IdlReprModifier = {
//   packed?: boolean;
//   align?: number;
// };

type IdlReprRust struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    #[serde(flatten)]
	//	    pub modifier: IdlReprModifier,
	IdlReprModifier
}

func (repr IdlReprRust) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		IdlReprModifier
	}{
		Kind:            "rust",
		IdlReprModifier: repr.IdlReprModifier,
	})
}

func (i *IdlReprRust) UnmarshalJSON(data []byte) error {
	//	err := tools.RequireFields(
	//		data,
	//		"kind",
	//	)
	//	if err != nil {
	//		return err
	//	}
	type Alias IdlReprRust
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "rust" {
		return fmt.Errorf("expected kind 'rust', got %s", alias.Kind)
	}
	i.Kind = "rust"
	i.IdlReprModifier = alias.IdlReprModifier
	return nil
}

func (IdlReprRust) _is_IdlRepr() {}

type IdlReprC struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    #[serde(flatten)]
	//	    pub modifier: IdlReprModifier,
	IdlReprModifier
}

func (repr IdlReprC) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		IdlReprModifier
	}{
		Kind:            "c",
		IdlReprModifier: repr.IdlReprModifier,
	})
}

func (i *IdlReprC) UnmarshalJSON(data []byte) error {
	type Alias IdlReprC
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "c" {
		return fmt.Errorf("expected kind 'c', got %s", alias.Kind)
	}
	i.Kind = "c"
	i.IdlReprModifier = alias.IdlReprModifier
	return nil
}

func (IdlReprC) _is_IdlRepr() {}

type IdlReprTransparent struct {
	//	    pub kind: String,
	Kind string `json:"kind"`
	//	    #[serde(flatten)]
	//	    pub modifier: IdlReprModifier,
	IdlReprModifier
}

func (repr IdlReprTransparent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		IdlReprModifier
	}{
		Kind:            "transparent",
		IdlReprModifier: repr.IdlReprModifier,
	})
}

func (i *IdlReprTransparent) UnmarshalJSON(data []byte) error {
	type Alias IdlReprTransparent
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "transparent" {
		return fmt.Errorf("expected kind 'transparent', got %s", alias.Kind)
	}
	i.Kind = "transparent"
	i.IdlReprModifier = alias.IdlReprModifier
	return nil
}

func (IdlReprTransparent) _is_IdlRepr() {}

// pub struct IdlReprModifier {
type IdlReprModifier struct {
	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub packed: bool,
	Packed bool `json:"packed,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub align: Option<usize>,
	Align Option[uint] `json:"align,omitzero"`

	//	}
}

func (IdlReprModifier) _is_IdlRepr() {}

func into_IdlRepr(
	dst *IdlRepr,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlRepr[*IdlReprRust],
		tryUnmarshal_IdlRepr[*IdlReprC],
		tryUnmarshal_IdlRepr[*IdlReprTransparent],
	)
}

func tryUnmarshal_IdlRepr[T IdlRepr](
	data []byte,
) (IdlRepr, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}
