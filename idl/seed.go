package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

// #[serde(tag = "kind", rename_all = "lowercase")]
//
//	pub enum IdlSeed {
//	    Const(IdlSeedConst),
//	    Arg(IdlSeedArg),
//	    Account(IdlSeedAccount),
//	}
type IdlSeed interface {
	_is_IdlSeed()
}

// pub struct IdlSeedConst {
type IdlSeedConst struct {
	//	    pub value: Vec<u8>,
	Value []byte `json:"value"`
	//	}
}

var _IdlSeedConst IdlSeed = (*IdlSeedConst)(nil)

func (IdlSeedConst) _is_IdlSeed() {}
func (c IdlSeedConst) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind  string `json:"kind"`
		Value []uint `json:"value"`
	}{
		Kind: "const",
		Value: func() []uint {
			if c.Value == nil {
				return nil
			}
			out := make([]uint, len(c.Value))
			for i, v := range c.Value {
				out[i] = uint(v)
			}
			return out
		}(),
	})
}

func (c *IdlSeedConst) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"value",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		//	    pub kind: String,
		Kind string `json:"kind"`
		//	    pub value: Vec<u8>,
		Value []uint `json:"value"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "const" {
		return fmt.Errorf("expected kind 'const', got %s", alias.Kind)
	}
	c.Value = func() []byte {
		if alias.Value == nil {
			return nil
		}
		out := make([]byte, len(alias.Value))
		for i, v := range alias.Value {
			out[i] = byte(v)
		}
		return out
	}()
	return nil
}

// pub struct IdlSeedArg {
type IdlSeedArg struct {
	//     pub path: String,
	Path string `json:"path"`
	// }
}

func (IdlSeedArg) _is_IdlSeed() {}
func (a IdlSeedArg) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind string `json:"kind"`
		Path string `json:"path"`
	}{
		Kind: "arg",
		Path: a.Path,
	})
}

func (a *IdlSeedArg) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"path",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Kind string `json:"kind"`
		//	    pub path: String,
		Path string `json:"path"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "arg" {
		return fmt.Errorf("expected kind 'arg', got %s", alias.Kind)
	}
	a.Path = alias.Path
	return nil
}

// pub struct IdlSeedAccount {
type IdlSeedAccount struct {
	//	    pub path: String,
	Path string `json:"path"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub account: Option<String>,
	Account Option[string] `json:"account,omitzero"`
	//	}
}

func (IdlSeedAccount) _is_IdlSeed() {}
func (a IdlSeedAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind    string         `json:"kind"`
		Path    string         `json:"path"`
		Account Option[string] `json:"account,omitzero"`
	}{
		Kind:    "account",
		Path:    a.Path,
		Account: a.Account,
	})
}

func (a *IdlSeedAccount) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"path",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Kind string `json:"kind"`
		//	    pub path: String,
		Path    string         `json:"path"`
		Account Option[string] `json:"account,omitzero"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Kind != "account" {
		return fmt.Errorf("expected kind 'account', got %s", alias.Kind)
	}
	a.Path = alias.Path
	a.Account = alias.Account
	return nil
}

func into_IdlSeed(
	seed *IdlSeed,
	data json.RawMessage,
) error {
	return tools.Into(
		seed,
		data,
		tryUnmarshal_IdlSeed[*IdlSeedConst],
		tryUnmarshal_IdlSeed[*IdlSeedArg],
		tryUnmarshal_IdlSeed[*IdlSeedAccount],
	)
}

func tryUnmarshal_IdlSeed[T IdlSeed](
	data []byte,
) (IdlSeed, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
