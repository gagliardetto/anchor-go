package idltype

import (
	"encoding/json"

	"github.com/gagliardetto/anchor-go/tools"
)

func Into(
	dst *IdlType,
	data []byte,
) error {
	return tools.Into(
		dst,
		data,
		tryUnmarshal[*Bool],
		tryUnmarshal[*U8],
		tryUnmarshal[*I8],
		tryUnmarshal[*U16],
		tryUnmarshal[*I16],
		tryUnmarshal[*U32],
		tryUnmarshal[*I32],
		tryUnmarshal[*F32],
		tryUnmarshal[*U64],
		tryUnmarshal[*I64],
		tryUnmarshal[*F64],
		tryUnmarshal[*U128],
		tryUnmarshal[*I128],
		tryUnmarshal[*U256],
		tryUnmarshal[*I256],
		tryUnmarshal[*Bytes],
		tryUnmarshal[*String],
		tryUnmarshal[*Pubkey],
		tryUnmarshal[*Option],
		tryUnmarshal[*COption],
		tryUnmarshal[*Vec],
		tryUnmarshal[*Array],
		tryUnmarshal[*Defined],
		tryUnmarshal[*Generic],
	)
}

func tryUnmarshal[T IdlType](data []byte) (IdlType, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}
