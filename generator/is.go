package generator

import "github.com/gagliardetto/anchor-go/idl/idltype"

func IsOption(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Option:
		return true
	default:
		return false
	}
}

func IsCOption(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.COption:
		return true
	default:
		return false
	}
}

func IsDefined(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Defined:
		return true
	default:
		return false
	}
}

func IsVec(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Vec:
		return true
	default:
		return false
	}
}

func IsArray(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Array:
		return true
	default:
		return false
	}
}

func IsIDLTypeKind(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Bool:
		return true
	case *idltype.U8:
		return true
	case *idltype.I8:
		return true
	case *idltype.U16:
		return true
	case *idltype.I16:
		return true
	case *idltype.U32:
		return true
	case *idltype.I32:
		return true
	case *idltype.F32:
		return true
	case *idltype.U64:
		return true
	case *idltype.I64:
		return true
	case *idltype.F64:
		return true
	case *idltype.U128:
		return true
	case *idltype.I128:
		return true
	case *idltype.U256:
		return true
	case *idltype.I256:
		return true
	case *idltype.Bytes:
		return true
	case *idltype.String:
		return true
	case *idltype.Pubkey:
		return true
	default:
		return false
	}
}

func IsBool(v idltype.IdlType) bool {
	switch v.(type) {
	case *idltype.Bool:
		return true
	default:
		return false
	}
}
