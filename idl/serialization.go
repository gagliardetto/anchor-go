package idl

import (
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/anchor-go/tools"
)

// export type IdlSerialization =
//   | "borsh"
//   | "bytemuck"
//   | "bytemuckunsafe"
//   | { custom: string };

// #[serde(rename_all = "lowercase")]
// #[non_exhaustive]
//
//	pub enum IdlSerialization {
//	    #[default]
//	    Borsh,
//	    Bytemuck,
//	    BytemuckUnsafe,
//	    Custom(String),
//	}
type IdlSerialization interface {
	_is_IdlSerialization()
}

type IdlSerializationBorsh struct {
	_ struct{}
}

func (IdlSerializationBorsh) _is_IdlSerialization() {}
func (IdlSerializationBorsh) MarshalJSON() ([]byte, error) {
	return []byte(`"borsh"`), nil
}

func (IdlSerializationBorsh) UnmarshalJSON(data []byte) error {
	if string(data) != `"borsh"` {
		return fmt.Errorf("expected 'borsh', got %s", string(data))
	}
	return nil
}

type IdlSerializationBytemuck struct {
	_ struct{}
}

func (IdlSerializationBytemuck) _is_IdlSerialization() {}
func (IdlSerializationBytemuck) MarshalJSON() ([]byte, error) {
	return []byte(`"bytemuck"`), nil
}

func (IdlSerializationBytemuck) UnmarshalJSON(data []byte) error {
	if string(data) != `"bytemuck"` {
		return fmt.Errorf("expected 'bytemuck', got %s", string(data))
	}
	return nil
}

type IdlSerializationBytemuckUnsafe struct {
	_ struct{}
}

func (IdlSerializationBytemuckUnsafe) _is_IdlSerialization() {}
func (IdlSerializationBytemuckUnsafe) MarshalJSON() ([]byte, error) {
	return []byte(`"bytemuckunsafe"`), nil
}

func (IdlSerializationBytemuckUnsafe) UnmarshalJSON(data []byte) error {
	if string(data) != `"bytemuckunsafe"` {
		return fmt.Errorf("expected 'bytemuckunsafe', got %s", string(data))
	}
	return nil
}

type IdlSerializationCustom struct {
	_      struct{}
	Custom string `json:"custom"`
}

func (IdlSerializationCustom) _is_IdlSerialization() {}
func (i IdlSerializationCustom) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"custom":"%s"}`, i.Custom)), nil
}

func (i *IdlSerializationCustom) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if len(data) < 10 || data[0] != '{' || data[len(data)-1] != '}' {
		return fmt.Errorf("expected object, got %s", string(data))
	}
	type Alias struct {
		Custom string `json:"custom"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	if alias.Custom == "" {
		return fmt.Errorf("expected 'custom' field, got %s", string(data))
	}
	i.Custom = alias.Custom
	return nil
}

func into_IdlSerialization(
	dst *IdlSerialization,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal_IdlSerialization[*IdlSerializationBorsh],
		tryUnmarshal_IdlSerialization[*IdlSerializationBytemuck],
		tryUnmarshal_IdlSerialization[*IdlSerializationBytemuckUnsafe],
		tryUnmarshal_IdlSerialization[*IdlSerializationCustom],
	)
}

func tryUnmarshal_IdlSerialization[T IdlSerialization](
	data []byte,
) (IdlSerialization, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}
