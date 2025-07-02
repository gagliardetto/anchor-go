package idl

import (
	"encoding/json"

	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

// pub struct IdlField {
type IdlField struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    #[serde(rename = "type")]
	//	    pub ty: IdlType,
	Ty idltype.IdlType `json:"type"`
	//	}
}

func (f *IdlField) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
		"type",
	)
	if err != nil {
		return err
	}

	type Alias struct {
		Name string          `json:"name"`
		Docs []string        `json:"docs,omitzero"`
		Ty   json.RawMessage `json:"type"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	f.Name = alias.Name
	f.Docs = alias.Docs
	{
		var ty idltype.IdlType
		err := idltype.Into(&ty, alias.Ty)
		if err != nil {
			return err
		}
		f.Ty = ty
	}

	return nil
}
