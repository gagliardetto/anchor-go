package idl

import (
	"encoding/json"

	"github.com/gagliardetto/anchor-go/tools"
)

// pub struct IdlPda {
type IdlPda struct {
	//	    pub seeds: Vec<IdlSeed>,
	Seeds []IdlSeed `json:"seeds"`

	//	    #[serde(skip_serializing_if = "is_default")]
	//	    pub program: Option<IdlSeed>,
	Program Option[IdlSeed] `json:"program,omitzero"`

	//	}
}

func (pda *IdlPda) UnmarshalJSON(data []byte) error {
	err := tools.RequireFields(
		data,
		"seeds",
	)
	if err != nil {
		return err
	}
	type Alias struct {
		//	    pub seeds: Vec<IdlSeed>,
		Seeds []json.RawMessage `json:"seeds"`
		//	    #[serde(skip_serializing_if = "is_default")]
		//	    pub program: Option<IdlSeed>,
		Program Option[json.RawMessage] `json:"program,omitzero"`
	}
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	// Unmarshal the seeds
	pda.Seeds = make([]IdlSeed, len(alias.Seeds))
	for i, seed := range alias.Seeds {
		var seedValue IdlSeed
		err := into_IdlSeed(&seedValue, seed)
		if err != nil {
			return err
		}
		pda.Seeds[i] = seedValue
	}
	// Unmarshal the program
	if alias.Program.IsSome() {
		var programValue IdlSeed
		err := into_IdlSeed(&programValue, alias.Program.Unwrap())
		if err != nil {
			return err
		}
		pda.Program = Some(programValue)
	} else {
		pda.Program = None[IdlSeed]()
	}

	return nil
}
