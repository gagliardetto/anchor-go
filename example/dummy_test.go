package example

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/encrypt-x/solana-anchor-go/generated/dummy"
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
	fmt.Printf("parsed %d instructions of dummy program\n", len(instructions))

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
	//events from log: <*>{Incremented <*>{AiCB6Pp7uiJDky3yg3zb462FMcD6GpMvP4bd3B1BQf5E LST1 100}}
	//parsed 1 instructions of dummy program
	//incremented LST1 token for 100 amount
}

//func Example2() {
//	dummy.NewInitializeInstruction()...
//	dummy.NewDecrementInstructionBuilder()..
//	dummy.NewIncrementInstructionBuilder()..
//}
