package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
	"github.com/gagliardetto/solana-go"
)

//	pub enum IdlInstructionAccountItem {
//	 Composite(IdlInstructionAccounts),
//	 Single(IdlInstructionAccount),
//	}
type IdlInstructionAccountItem interface {
	_is_IdlInstructionAccountItem()
}

// pub struct IdlInstructionAccount {
type IdlInstructionAccount struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub writable: bool,
	Writable bool `json:"writable,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub signer: bool,
	Signer bool `json:"signer,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub optional: bool,
	Optional bool `json:"optional,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub address: Option<String>,
	Address Option[solana.PublicKey] `json:"address,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub pda: Option<IdlPda>,
	Pda Option[IdlPda] `json:"pda,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub relations: Vec<String>,
	Relations []string `json:"relations,omitzero"`
	//	}
}

func (i *IdlInstructionAccount) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
	)
	if err != nil {
		return err
	}
	// err = tools.RequireOneOfFields(
	// 	data,
	// 	"address",
	// 	"pda",
	// 	"signer",
	// 	"writable",
	// )
	// if err != nil {
	// 	return err
	// }
	type Alias IdlInstructionAccount
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*i = IdlInstructionAccount(alias)
	return nil
}

func (i *IdlInstructionAccount) _is_IdlInstructionAccountItem() {}

// pub struct IdlInstructionAccounts {
type IdlInstructionAccounts struct {
	//	    pub name: String,
	Name string `json:"name"`
	//	    pub accounts: Vec<IdlInstructionAccountItem>,
	Accounts []IdlInstructionAccountItem `json:"accounts"`
	//	}
}

func (ia *IdlInstructionAccounts) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
		"accounts",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Name     string            `json:"name"`
		Accounts []json.RawMessage `json:"accounts"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	// Unmarshal the accounts into the correct types.
	ia.Name = alias.Name
	ia.Accounts = make([]IdlInstructionAccountItem, len(alias.Accounts))
	for i, account := range alias.Accounts {
		err := into_IdlInstructionAccountItem(
			&ia.Accounts[i],
			account,
		)
		if err != nil {
			return fmt.Errorf("failed to unmarshal account %d: %w", i, err)
		}
	}
	return nil
}

func (i *IdlInstructionAccounts) _is_IdlInstructionAccountItem() {}

func into_IdlInstructionAccountItem(
	dst *IdlInstructionAccountItem,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlInstructionAccountItem[*IdlInstructionAccounts],
		tryUnmarshal_IdlInstructionAccountItem[*IdlInstructionAccount],
	)
}

func tryUnmarshal_IdlInstructionAccountItem[T IdlInstructionAccountItem](
	data []byte,
) (IdlInstructionAccountItem, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}
