package generator

import (
	"testing"

	. "github.com/dave/jennifer/jen"
	"github.com/gagliardetto/anchor-go/idl"
	"github.com/gagliardetto/anchor-go/idl/idltype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderGeneratedCode(t *testing.T, code Code) string {
	t.Helper()
	file := NewFile("testpkg")
	file.Add(code)
	return file.GoString()
}

func TestGenInstructionTypeGeneratesGetters(t *testing.T) {
	gen := NewGenerator(&idl.Idl{}, &GeneratorOptions{Package: "testpkg"})

	code, err := gen.gen_instructionType(idl.IdlInstruction{
		Name: "transfer",
		Args: []idl.IdlField{
			{Name: "amount", Ty: &idltype.U64{}},
			{Name: "memo", Ty: &idltype.Option{Option: &idltype.String{}}},
		},
		Accounts: []idl.IdlInstructionAccountItem{
			&idl.IdlInstructionAccount{Name: "source", Writable: true, Signer: true, Optional: true},
		},
	})
	require.NoError(t, err)

	generated := renderGeneratedCode(t, code)
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetName() string")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetAmount() uint64")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetMemo() string")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetMemoPtr() *string")
	assert.Contains(t, generated, "func (obj *TransferInstruction) HasMemo() bool")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetSource() solanago.PublicKey")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetSourceWritable() bool")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetSourceSigner() bool")
	assert.Contains(t, generated, "func (obj *TransferInstruction) GetSourceOptional() bool")
}

func TestGenInstructionTypeGetterCollisionUsesFieldSuffix(t *testing.T) {
	gen := NewGenerator(&idl.Idl{}, &GeneratorOptions{Package: "testpkg"})

	code, err := gen.gen_instructionType(idl.IdlInstruction{
		Name: "collision",
		Args: []idl.IdlField{
			{Name: "discriminator", Ty: &idltype.Option{Option: &idltype.String{}}},
		},
	})
	require.NoError(t, err)

	generated := renderGeneratedCode(t, code)
	assert.Contains(t, generated, "func (obj *CollisionInstruction) GetDiscriminatorField() string")
	assert.Contains(t, generated, "func (obj *CollisionInstruction) GetDiscriminatorFieldPtr() *string")
	assert.NotContains(t, generated, "func (obj *CollisionInstruction) GetDiscriminator() string")
}

func TestGenEventTypeGeneratesFieldGetters(t *testing.T) {
	gen := NewGenerator(&idl.Idl{
		Types: idl.IdTypeDef_slice{
			{
				Name: "transfer_event",
				Ty: &idl.IdlTypeDefTyStruct{
					Kind: "struct",
					Fields: idl.IdlDefinedFieldsNamed{
						{Name: "amount", Ty: &idltype.U64{}},
						{Name: "authority", Ty: &idltype.Option{Option: &idltype.Pubkey{}}},
					},
				},
			},
		},
	}, &GeneratorOptions{Package: "testpkg"})

	code, err := gen.gen_eventType(idl.IdlEvent{Name: "transfer_event"})
	require.NoError(t, err)

	generated := renderGeneratedCode(t, code)
	assert.Contains(t, generated, "func (obj *TransferEvent) GetName() string")
	assert.Contains(t, generated, "func (obj *TransferEvent) GetAmount() uint64")
	assert.Contains(t, generated, "func (obj *TransferEvent) GetAuthority() solanago.PublicKey")
	assert.Contains(t, generated, "func (obj *TransferEvent) GetAuthorityPtr() *solanago.PublicKey")
	assert.Contains(t, generated, "func (obj *TransferEvent) HasAuthority() bool")
}

func TestGenEventParserUsesTypedHelper(t *testing.T) {
	gen := NewGenerator(&idl.Idl{}, &GeneratorOptions{Package: "testpkg"})

	code, err := gen.gen_eventParser([]string{"TransferEvent"})
	require.NoError(t, err)

	generated := renderGeneratedCode(t, code)
	assert.Contains(t, generated, "func ParseAnyEvent(eventData []byte) (any, error)")
	assert.Contains(t, generated, "func ParseEventTyped[T Event](eventData []byte) (T, error)")
	assert.Contains(t, generated, "return ParseEventTyped[*TransferEvent](eventData)")

	// emit_cpi compatibility surface: constant + dedicated helper.
	assert.Contains(t, generated, "AnchorEmitCPIDiscriminator = [8]byte{228, 69, 165, 46, 81, 203, 154, 29}")
	assert.Contains(t, generated, "func ParseEmitCPIEvent(data []byte) (Event, error)")
	assert.Contains(t, generated, "!bytes.Equal(data[:8], AnchorEmitCPIDiscriminator[:])")
	assert.Contains(t, generated, "return ParseEvent(data[8:])")
}

func TestGenInstructionParserStripsEmitCPIInParseParsable(t *testing.T) {
	gen := NewGenerator(&idl.Idl{}, &GeneratorOptions{Package: "testpkg"})

	code, err := gen.gen_instructionParser([]string{"FooInstruction"}, []string{"foo"})
	require.NoError(t, err)

	generated := renderGeneratedCode(t, code)
	// ParseParsable should short-circuit on the emit_cpi sentinel before
	// attempting the instruction/event fallback chain.
	assert.Contains(t, generated, "func ParseParsable(data []byte) (Parsable, error)")
	assert.Contains(t, generated, "len(data) >= 8 && bytes.Equal(data[:8], AnchorEmitCPIDiscriminator[:])")
	assert.Contains(t, generated, "return ParseEvent(data[8:])")
}
