package idl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/gagliardetto/anchor-go/tools"
	"github.com/gagliardetto/solana-go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// From https://github.com/solana-foundation/anchor/blob/8b0e965c65fb96b6865be53c478a16007984a566/idl/spec/src/lib.rs

func ParseFromFilepath(path string) (*Idl, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if IsOldIdl(data) {
		return nil, fmt.Errorf("the IDL at '%s' is an old IDL format, please use the anchor cli to convert the old idl to the new format with `anchor idl convert <PATH_TO_IDL_JSON>`, or use the latest version of the Anchor CLI to generate a new IDL", path)
	}
	return Parse(data)
}

func Parse(data []byte) (*Idl, error) {
	var idl Idl
	err := json.Unmarshal(data, &idl)
	if err != nil {
		return nil, err
	}
	return &idl, nil
}

// pub struct Idl {
type Idl struct {
	//     pub address: String,
	Address *solana.PublicKey `json:"address,omitzero"`

	//	    pub metadata: IdlMetadata,
	Metadata IdlMetadata `json:"metadata,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    pub instructions: Vec<IdlInstruction>,
	Instructions []IdlInstruction `json:"instructions"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub accounts: Vec<IdlAccount>,
	Accounts []IdlAccount `json:"accounts,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub events: Vec<IdlEvent>,
	Events []IdlEvent `json:"events,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub errors: Vec<IdlErrorCode>,
	Errors []IdlErrorCode `json:"errors,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub types: Vec<IdlTypeDef>,
	Types IdTypeDef_slice `json:"types,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub constants: Vec<IdlConst>,
	Constants []IdlConst `json:"constants,omitzero"`
	//	}
}

type IdTypeDef_slice []IdlTypeDef

// ByName returns the IdlTypeDef with the given name, or nil if not found.
func (slice IdTypeDef_slice) ByName(name string) *IdlTypeDef {
	for i := range slice {
		if slice[i].Name == name {
			return &slice[i]
		}
	}
	return nil
}

func IsOldIdl(raw []byte) bool {
	// if contains "isMut", "isSigner"
	return bytes.Contains(raw, []byte("isMut")) ||
		bytes.Contains(raw, []byte("isSigner")) || bytes.HasPrefix(raw, []byte("{\"version\":"))
}

// Idl.UnmarshalJSON
func (i *Idl) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"instructions",
		// "metadata",
	)
	if err != nil {
		return err
	}
	{
		got := gjson.GetBytes(data, "address")
		if !got.Exists() || got.Type != gjson.String || got.Value().(string) == "" {
			data, err = sjson.DeleteBytes(data, "address")
		} else {
			// If the address is not a valid public key, we should remove it.
			if _, err := solana.PublicKeyFromBase58(got.Value().(string)); err != nil {
				return fmt.Errorf("invalid address in IDL: %q is not a valid public key: %w", got.Value().(string), err)
			}
		}
	}
	type Alias Idl
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*i = Idl(alias)
	return nil
}

// pub struct IdlMetadata {
type IdlMetadata struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    pub version: String,
	Version string `json:"version"`

	//	    pub spec: String,
	Spec string `json:"spec"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub description: Option<String>,
	Description Option[string] `json:"description,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub repository: Option<String>,
	Repository Option[string] `json:"repository,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub dependencies: Vec<IdlDependency>,
	Dependencies []IdlDependency `json:"dependencies,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub contact: Option<String>,
	Contact Option[string] `json:"contact,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub deployments: Option<IdlDeployments>,
	Deployments Option[IdlDeployments] `json:"deployments,omitzero"`
	//	}
}

// pub struct IdlDependency {
type IdlDependency struct {
	//     pub name: String,
	Name string `json:"name"`
	//     pub version: String,
	Version string `json:"version"`
	// }
}

// pub struct IdlDeployments {
type IdlDeployments struct {
	//     pub mainnet: Option<String>,
	Mainnet Option[string] `json:"mainnet"`
	//     pub testnet: Option<String>,
	Testnet Option[string] `json:"testnet"`
	//     pub devnet: Option<String>,
	Devnet Option[string] `json:"devnet"`
	//     pub localnet: Option<String>,
	Localnet Option[string] `json:"localnet"`
	// }
}

// pub struct IdlAccount {
type IdlAccount struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    pub discriminator: IdlDiscriminator,
	Discriminator IdlDiscriminator `json:"discriminator"`

	//	}
}

// pub struct IdlEvent {
type IdlEvent struct {
	//     pub name: String,
	Name string `json:"name"`
	//     pub discriminator: IdlDiscriminator,
	Discriminator IdlDiscriminator `json:"discriminator"`
	// }
}

// pub struct IdlConst {
type IdlConst struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    #[serde(rename = "type")]
	//	    pub ty: IdlType,
	Ty idltype.IdlType `json:"type"`

	//	    pub value: String,
	Value string `json:"value"`
	//	}
}

func (i *IdlConst) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Name  string          `json:"name"`
		Docs  []string        `json:"docs,omitzero"`
		Ty    json.RawMessage `json:"type"`
		Value string          `json:"value"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	i.Name = alias.Name
	i.Docs = alias.Docs
	{
		var ty idltype.IdlType
		err := idltype.Into(&ty, alias.Ty)
		if err != nil {
			return err
		}
		i.Ty = ty
	}
	i.Value = alias.Value
	return nil
}

// pub struct IdlErrorCode {
type IdlErrorCode struct {
	//	    pub code: u32,
	Code uint32 `json:"code"`

	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub msg: Option<String>,
	Msg Option[string] `json:"msg,omitzero"`

	//	}
}

// pub struct IdlTypeDef {
type IdlTypeDef struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub docs: Vec<String>,
	Docs []string `json:"docs,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub serialization: IdlSerialization,
	Serialization IdlSerialization `json:"serialization,omitzero"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub repr: Option<IdlRepr>,
	Repr Option[IdlRepr] `json:"repr,omitzero"`

	//	    #[serde(default, skip_serializing_if = "is_default")]
	//	    pub generics: Vec<IdlTypeDefGeneric>,
	Generics []IdlTypeDefGeneric `json:"generics,omitzero"`

	//	    #[serde(rename = "type")]
	//	    pub ty: IdlTypeDefTy,
	Ty IdlTypeDefTy `json:"type"`
	//	}
}

func (i *IdlTypeDef) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
		// "serialization",
		"type",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Name          string            `json:"name"`
		Docs          []string          `json:"docs,omitzero"`
		Serialization json.RawMessage   `json:"serialization"`
		Repr          json.RawMessage   `json:"repr,omitzero"`
		Generics      []json.RawMessage `json:"generics,omitzero"`
		Ty            json.RawMessage   `json:"type"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	i.Name = alias.Name
	i.Docs = alias.Docs
	if len(alias.Serialization) > 0 {
		var serialization IdlSerialization
		err = into_IdlSerialization(&serialization, alias.Serialization)
		if err != nil {
			return err
		}
		i.Serialization = serialization
	}
	if len(alias.Repr) > 0 {
		var repr IdlRepr
		err = into_IdlRepr(&repr, alias.Repr)
		if err != nil {
			return err
		}
		i.Repr = Some(repr)
	}
	if len(alias.Generics) > 0 {
		generics := make([]IdlTypeDefGeneric, len(alias.Generics))
		for i, raw := range alias.Generics {
			var generic IdlTypeDefGeneric
			err = into_IdlTypeDefGeneric(&generic, raw)
			if err != nil {
				return err
			}
			generics[i] = generic
		}
		i.Generics = generics
	}
	if len(alias.Ty) > 0 {
		var ty IdlTypeDefTy
		err = into_IdlTypeDefTy(&ty, alias.Ty)
		if err != nil {
			return err
		}
		i.Ty = ty
	}
	return nil
}

// pub struct IdlEnumVariant {
type IdlEnumVariant struct {
	//	    pub name: String,
	Name string `json:"name"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub fields: Option<IdlDefinedFields>,
	Fields Option[IdlDefinedFields] `json:"fields,omitzero"`

	//	}
}

func (variant *IdlEnumVariant) IsSimple() bool {
	// it's a simple uint8 if there is no fields data
	return variant.Fields.IsNone() || IsNil(variant.Fields)
}

func (i *IdlEnumVariant) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"name",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		Name   string          `json:"name"`
		Fields json.RawMessage `json:"fields,omitzero"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	i.Name = alias.Name
	if len(alias.Fields) > 0 {
		var fields IdlDefinedFields
		err = into_IdlDefinedFields(&fields, alias.Fields)
		if err != nil {
			return err
		}
		i.Fields = Some(fields)
	}
	return nil
}

// #[serde(rename_all = "lowercase")]
//
//	pub enum IdlArrayLen {
//	    Generic(String),
//	    #[serde(untagged)]
//	    Value(usize),
//	}
type IdlArrayLen interface {
	_is_IdlArrayLen()
}

// #[serde(rename_all = "lowercase")]
// #[non_exhaustive]
//
//	pub enum IdlType {
//	    Bool,
//	    U8,
//	    I8,
//	    U16,
//	    I16,
//	    U32,
//	    I32,
//	    F32,
//	    U64,
//	    I64,
//	    F64,
//	    U128,
//	    I128,
//	    U256,
//	    I256,
//	    Bytes,
//	    String,
//	    Pubkey,
//	    Option(Box<IdlType>),
//	    Vec(Box<IdlType>),
//	    Array(Box<IdlType>, IdlArrayLen),
//	    Defined {
//	        name: String,
//	        #[serde(default, skip_serializing_if = "is_default")]
//	        generics: Vec<IdlGenericArg>,
//	    },
//	    Generic(String),
//	}

type Array[T any, L IdlArrayLen] struct {
	Inner  T
	Length L
}

type Defined struct {
	Name     string                  `json:"name"`
	Generics []idltype.IdlGenericArg `json:"generics,omitzero"`
}

type Generic string

// impl FromStr for IdlType {
//     type Err = anyhow::Error;

//     fn from_str(s: &str) -> Result<Self, Self::Err> {
//         let mut s = s.to_owned();
//         s.retain(|c| !c.is_whitespace());

//         let r = match s.as_str() {
//             "bool" => IdlType::Bool,
//             "u8" => IdlType::U8,
//             "i8" => IdlType::I8,
//             "u16" => IdlType::U16,
//             "i16" => IdlType::I16,
//             "u32" => IdlType::U32,
//             "i32" => IdlType::I32,
//             "f32" => IdlType::F32,
//             "u64" => IdlType::U64,
//             "i64" => IdlType::I64,
//             "f64" => IdlType::F64,
//             "u128" => IdlType::U128,
//             "i128" => IdlType::I128,
//             "u256" => IdlType::U256,
//             "i256" => IdlType::I256,
//             "Vec<u8>" => IdlType::Bytes,
//             "String" | "&str" | "&'staticstr" => IdlType::String,
//             "Pubkey" => IdlType::Pubkey,
//             _ => {
//                 if let Some(inner) = s.strip_prefix("Option<") {
//                     let inner_ty = Self::from_str(
//                         inner
//                             .strip_suffix('>')
//                             .ok_or_else(|| anyhow!("Invalid Option"))?,
//                     )?;
//                     return Ok(IdlType::Option(Box::new(inner_ty)));
//                 }

//                 if let Some(inner) = s.strip_prefix("Vec<") {
//                     let inner_ty = Self::from_str(
//                         inner
//                             .strip_suffix('>')
//                             .ok_or_else(|| anyhow!("Invalid Vec"))?,
//                     )?;
//                     return Ok(IdlType::Vec(Box::new(inner_ty)));
//                 }

//                 if s.starts_with('[') {
//                     fn array_from_str(inner: &str) -> IdlType {
//                         match inner.strip_suffix(']') {
//                             Some(nested_inner) => array_from_str(&nested_inner[1..]),
//                             None => {
//                                 let (raw_type, raw_length) = inner.rsplit_once(';').unwrap();
//                                 let ty = IdlType::from_str(raw_type).unwrap();
//                                 let len = match raw_length.replace('_', "").parse::<usize>() {
//                                     Ok(len) => IdlArrayLen::Value(len),
//                                     Err(_) => IdlArrayLen::Generic(raw_length.to_owned()),
//                                 };
//                                 IdlType::Array(Box::new(ty), len)
//                             }
//                         }
//                     }
//                     return Ok(array_from_str(&s));
//                 }

//                 // Defined
//                 let (name, generics) = if let Some(i) = s.find('<') {
//                     (
//                         s.get(..i).unwrap().to_owned(),
//                         s.get(i + 1..)
//                             .unwrap()
//                             .strip_suffix('>')
//                             .unwrap()
//                             .split(',')
//                             .map(|g| g.trim().to_owned())
//                             .map(|g| {
//                                 if g.parse::<bool>().is_ok()
//                                     || g.parse::<u128>().is_ok()
//                                     || g.parse::<i128>().is_ok()
//                                     || g.parse::<char>().is_ok()
//                                 {
//                                     Ok(IdlGenericArg::Const { value: g })
//                                 } else {
//                                     Self::from_str(&g).map(|ty| IdlGenericArg::Type { ty })
//                                 }
//                             })
//                             .collect::<Result<Vec<_>, _>>()?,
//                     )
//                 } else {
//                     (s.to_owned(), vec![])
//                 };

//	                IdlType::Defined { name, generics }
//	            }
//	        };
//	        Ok(r)
//	    }
//	}
// TODO: uncomment.
// func _IdlType_from_str(s string) (IdlType, error) {
// 	s = strings.TrimSpace(s)

// 	switch s {
// 	case "bool":
// 		return IdlTypeBool{}, nil
// 	case "u8":
// 		return IdlTypeU8{}, nil
// 	case "i8":
// 		return IdlTypeI8{}, nil
// 	case "u16":
// 		return IdlTypeU16{}, nil
// 	case "i16":
// 		return IdlTypeI16{}, nil
// 	case "u32":
// 		return IdlTypeU32{}, nil
// 	case "i32":
// 		return IdlTypeI32{}, nil
// 	case "f32":
// 		return IdlTypeF32{}, nil
// 	case "u64":
// 		return IdlTypeU64{}, nil
// 	case "i64":
// 		return IdlTypeI64{}, nil
// 	case "f64":
// 		return IdlTypeF64{}, nil
// 	case "u128":
// 		return IdlTypeU128{}, nil
// 	case "i128":
// 		return IdlTypeI128{}, nil
// 	case "u256":
// 		return IdlTypeU256{}, nil
// 	case "i256":
// 		return IdlTypeI256{}, nil
// 	case "Vec<u8>":
// 		return IdlTypeBytes{}, nil
// 	case "String", "&str", "&'staticstr":
// 		return IdlTypeString{}, nil
// 	case "Pubkey":
// 		return IdlTypePubkey{}, nil
// 	default:
// 		if strings.HasPrefix(s, "Option<") {
// 			s = strings.TrimPrefix(s, "Option<")
// 			s = strings.TrimSuffix(s, ">")
// 			innerTy, err := _IdlType_from_str(s)
// 			if err != nil {
// 				return nil, err
// 			}
// 			return &IdlTypeOption{Inner: innerTy}, nil
// 		}

// 		if strings.HasPrefix(s, "Vec<") {
// 			s = strings.TrimPrefix(s, "Vec<")
// 			s = strings.TrimSuffix(s, ">")
// 			innerTy, err := _IdlType_from_str(s)
// 			if err != nil {
// 				return nil, err
// 			}
// 			return &IdlTypeVec{Inner: innerTy}, nil
// 		}

// 		if strings.HasPrefix(s, "[") {
// 			var arrayLen IdlArrayLen

// 			for i := 0; i < len(s); i++ {
// 				if s[i] == ']' {
// 					break
// 				}
// 			}
// 			inner := s[1:i]
// 			s = s[i+1:]
// 			rawType, rawLength := strings.SplitN(inner, ";", 2)
// 			ty, err := _IdlType_from_str(rawType)
// 			if err != nil {
// 				return nil, err
// 			}
// 			rawLength = strings.ReplaceAll(rawLength, "_", "")
// 			len, err := strconv.ParseUint(rawLength, 10, 64)
// 			if err != nil {
// 				arrayLen = &IdlArrayLenGeneric{Value: rawLength}
// 			} else {
// 				arrayLen = &IdlArrayLenValue{Value: uint(len)}
// 			}
// 			return &IdlTypeArray{Inner: ty, Length: arrayLen}, nil
// 		}
// 		// Defined
// 		name := s
// 		generics := []IdlGenericArg{}
// 		if strings.Contains(s, "<") {
// 			i := strings.Index(s, "<")
// 			name = s[:i]
// 			genericsStr := s[i+1 : len(s)-1]
// 			genericsParts := strings.Split(genericsStr, ",")
// 			for _, g := range genericsParts {
// 				g = strings.TrimSpace(g)
// 				if _, err := strconv.ParseBool(g); err == nil {
// 					generics = append(generics, &IdlGenericArgConst{Value: g})
// 				} else if _, err := strconv.ParseUint(g, 10, 64); err == nil {
// 					generics = append(generics, &IdlGenericArgConst{Value: g})
// 				} else if _, err := strconv.ParseInt(g, 10, 64); err == nil {
// 					generics = append(generics, &IdlGenericArgConst{Value: g})
// 				} else if _, err := strconv.Unquote(g); err == nil {
// 					generics = append(generics, &IdlGenericArgConst{Value: g})
// 				} else {
// 					innerTy, err := _IdlType_from_str(g)
// 					if err != nil {
// 						return nil, err
// 					}
// 					generics = append(generics, &IdlGenericArgType{Inner: innerTy})
// 				}
// 			}
// 		}
// 		return &IdlTypeDefined{Name: name, Generics: generics}, nil
// 	}
// 	return nil, fmt.Errorf("invalid type: %s", s)
// }

// pub type IdlDiscriminator = Vec<u8>;
type IdlDiscriminator []byte

// IsEmpty returns true if the IdlDiscriminator is empty.
func (di IdlDiscriminator) IsEmpty() bool {
	return len(di) == 0
}

// MarshalJSON
func (di IdlDiscriminator) MarshalJSON() ([]byte, error) {
	if len(di) == 0 {
		return []byte("[]"), nil
	}
	asNumberArray := make([]uint, len(di))
	for i, b := range di {
		asNumberArray[i] = uint(b)
	}
	return json.Marshal(asNumberArray)
}

// UnmarshalJSON
func (di *IdlDiscriminator) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*di = nil
		return nil
	}
	var asNumberArray []uint
	if err := json.Unmarshal(data, &asNumberArray); err != nil {
		return err
	}
	*di = make(IdlDiscriminator, len(asNumberArray))
	for i, b := range asNumberArray {
		(*di)[i] = byte(b)
	}
	return nil
}

// /// Get whether the given data is the default of its type.
//
//	fn is_default<T: Default + PartialEq>(it: &T) -> bool {
//	    *it == T::default()
//	}
func is_default[T comparable](it T) bool {
	return it == *new(T)
}
