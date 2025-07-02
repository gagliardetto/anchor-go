package idl

import (
	"encoding/json"

	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
	bin "github.com/gagliardetto/binary"
)

// pub struct IdlInstruction {
type IdlInstruction struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    pub discriminator: IdlDiscriminator,
	Discriminator IdlDiscriminator `json:"discriminator"`

	//	    pub accounts: Vec<IdlInstructionAccountItem>,
	Accounts []IdlInstructionAccountItem `json:"accounts"`

	Args []IdlField `json:"args"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub returns: Option<IdlType>,
	Returns Option[idltype.IdlType] `json:"returns,omitzero"`
	//	}
}

func (ix *IdlInstruction) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
		"accounts",
		"args",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Name          string                  `json:"name"`
		Docs          []string                `json:"docs,omitzero"`
		Discriminator IdlDiscriminator        `json:"discriminator"`
		Accounts      []json.RawMessage       `json:"accounts"`
		Args          []IdlField              `json:"args"`
		Returns       Option[json.RawMessage] `json:"returns,omitzero"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	ix.Name = alias.Name
	ix.Docs = alias.Docs
	ix.Discriminator = alias.Discriminator
	{
		ix.Accounts = make([]IdlInstructionAccountItem, len(alias.Accounts))
		for i, raw := range alias.Accounts {
			err = into_IdlInstructionAccountItem(&ix.Accounts[i], raw)
			if err != nil {
				return err
			}
		}
	}
	ix.Args = alias.Args
	if alias.Returns.IsSome() {
		var returns idltype.IdlType
		err = idltype.Into(&returns, alias.Returns.Unwrap())
		if err != nil {
			return err
		}
		ix.Returns = Some(returns)
	} else {
		ix.Returns = None[idltype.IdlType]()
	}
	return nil
}

func (i *IdlInstruction) ComputeDiscriminator() (out [8]byte) {
	discrim := bin.SighashInstruction(i.Name)
	copy(out[:], discrim[:8])
	return out
}
