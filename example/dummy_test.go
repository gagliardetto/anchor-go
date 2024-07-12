package example

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/encrypt-x/solana-anchor-go/generated/dummy"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

//go:embed tx.json
var resJSON []byte

// an example of parsing a transaction.
// generate dummy IDL code first.
// $ go build && ./solana-anchor-go -src=./idl/fragmetric/dummy.json -pkg=dummy -dst=./generated/dummy
func Example1() {
	var res rpc.GetTransactionResult
	err := json.Unmarshal(resJSON, &res)
	if err != nil {
		panic(fmt.Errorf("cannot parse json - %v", err))
	}

	tx, err := res.Transaction.GetTransaction()
	if err != nil {
		panic(fmt.Errorf("cannot get tx from res - %v", err))
	}

	// set program id
	programID := solana.MustPublicKeyFromBase58("5yYKAKV5r62ooXrKZNpxr9Bkk7CTtpyJ8sXD7k2WryUc")
	dummy.SetProgramID(programID)
	fmt.Printf("dummy.ProgramID=%v, programID=%v\n", dummy.ProgramID, programID)

	// parsing events
	events, err := dummy.DecodeEvents(res.Meta.LogMessages)
	if err != nil {
		panic(fmt.Errorf("cannot get events from res - %v", err))
	}
	for _, evt := range events {
		spew.Printf("events from log: %v\n", evt)
	}

	// parsing instructions
	instructions, err := dummy.DecodeInstructions(&tx.Message)
	if err != nil {
		panic(fmt.Errorf("cannot decode ins - %v", err))
	}
	spew.Printf("parsed %d instructions of dummy program\n%v\n", len(instructions), instructions)

	for _, ins := range instructions {
		switch ins := ins.Impl.(type) {
		case *dummy.Increment:
			fmt.Printf("incremented %s token for %d amount\n", ins.Data.Token, ins.Data.Amount)
		case *dummy.Decrement:
			fmt.Printf("decremented %s token for %d amount\n", ins.Data.Token, ins.Data.Amount)
		default:
			fmt.Printf("dummy program's unknown instruction - %T\n", ins)
		}
	}

	//output:
	//dummy.ProgramID=5yYKAKV5r62ooXrKZNpxr9Bkk7CTtpyJ8sXD7k2WryUc, programID=5yYKAKV5r62ooXrKZNpxr9Bkk7CTtpyJ8sXD7k2WryUc
	//events from log: <*>{Incremented <*>{GtioYUL1PGYecn1yrgs89YyirzrftJn52gumLEeoU1qa LST1001 4361}}
	//parsed 1 instructions of dummy program
	//[<*>{{[11 18 104 9 104 174 59 33] <*>{<*>{LST1001 1111} [<*>{GtioYUL1PGYecn1yrgs89YyirzrftJn52gumLEeoU1qa true false} <*>{91zBeWL8kHBaMtaVrHwWsck1UacDKvje82QQ3HE2k8mJ true true} <*>{CRknqRhj9BfCcZTFEQfukeZiRxxmzd9MggiymG9ft3Jc true false}]}}}]
	//incremented LST1001 token for 1111 amount
}

//func Example2() {
//	dummy.NewInitializeInstruction()...
//	dummy.NewDecrementInstructionBuilder()..
//	dummy.NewIncrementInstructionBuilder()..
//}
