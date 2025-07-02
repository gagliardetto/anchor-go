package idl

import (
	"encoding/json"

	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
)

// #[serde(untagged)]
//
//	pub enum IdlDefinedFields {
//	    Named(Vec<IdlField>),
//	    Tuple(Vec<IdlType>),
//	}
type IdlDefinedFields interface {
	_is_IdlDefinedFields()
}

// export type IdlDefinedFields = IdlDefinedFieldsNamed | IdlDefinedFieldsTuple;

// export type IdlDefinedFieldsNamed = IdlField[];

// export type IdlDefinedFieldsTuple = IdlType[];

type IdlDefinedFieldsNamed []IdlField

func (IdlDefinedFieldsNamed) _is_IdlDefinedFields() {}

// func (f IdlDefinedFieldsNamed) MarshalJSON() ([]byte, error) {
// 	type Alias []IdlField
// 	alias := make(Alias, 0, len(f))
// 	for _, field := range f {
// 		alias = append(alias, field)
// 	}
// 	return json.Marshal(alias)
// }

func (f *IdlDefinedFieldsNamed) UnmarshalJSON(data []byte) error {
	type alias []IdlField
	if err := json.Unmarshal(data, (*alias)(f)); err != nil {
		return err
	}
	return nil
}

type IdlDefinedFieldsTuple []idltype.IdlType

func (IdlDefinedFieldsTuple) _is_IdlDefinedFields() {}

// func (f IdlDefinedFieldsTuple) MarshalJSON() ([]byte, error) {
// 	type Alias []idltype.IdlType
// 	alias := make(Alias, 0, len(f))
// 	for _, field := range f {
// 		alias = append(alias, field)
// 	}
// 	return json.Marshal(alias)
// }

func (f *IdlDefinedFieldsTuple) UnmarshalJSON(data []byte) error {
	type Alias []json.RawMessage
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	// Convert each raw message to IdlType
	converted := make([]idltype.IdlType, len(alias))
	for i, raw := range alias {
		var ty idltype.IdlType
		if err := idltype.Into(&ty, raw); err != nil {
			return err
		}
		converted[i] = ty
	}
	*f = converted
	return nil
}

func into_IdlDefinedFields(
	f *IdlDefinedFields,
	data json.RawMessage,
) error {
	return tools.Into(
		f,
		data,
		tryUnmarshal_IdlDefinedFields[IdlDefinedFieldsNamed],
		tryUnmarshal_IdlDefinedFields[IdlDefinedFieldsTuple],
	)
}

func tryUnmarshal_IdlDefinedFields[T IdlDefinedFields](
	data []byte,
) (IdlDefinedFields, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}
