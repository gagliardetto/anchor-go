package program

import "github.com/gagliardetto/solana-go"

// IDLAddress Deterministic IDL address as a function of the program id.
func IDLAddress(programId solana.PublicKey) (solana.PublicKey, error) {
	base, _, err := solana.FindProgramAddress(
		[][]byte{},
		programId,
	)
	if err != nil {
		return base, err
	}
	return solana.CreateWithSeed(base, seed(), programId)
}

// Seed for generating the idlAddress.
func seed() string {
	return "anchor:idl"
}
