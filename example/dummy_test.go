package example

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/encrypt-x/solana-anchor-go/generated/dummy"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

//go:embed dummy_tx.json
var txJSON []byte

//go:embed dummy_err.json
var errJSON []byte

var dummyProgramID = solana.MustPublicKeyFromBase58("A58NQYmJCyDPsc1EfaQZ99piFopPtCYArP242rLTbYbV")

// an example of parsing a transaction.
// generate dummy IDL code first.
// $ go build && ./solana-anchor-go -src=./idl/fragmetric/dummy.json -pkg=dummy -dst=./generated/dummy
func Example1() {
	var res rpc.GetTransactionResult
	err := json.Unmarshal(txJSON, &res)
	if err != nil {
		panic(fmt.Errorf("cannot parse json - %v", err))
	}

	tx, err := res.Transaction.GetTransaction()
	if err != nil {
		panic(fmt.Errorf("cannot get tx from res - %v", err))
	}

	// set program id
	dummy.SetProgramID(dummyProgramID)
	fmt.Printf("dummy.ProgramID=%v, programID=%v\n", dummy.ProgramID, dummyProgramID)

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
	//dummy.ProgramID=A58NQYmJCyDPsc1EfaQZ99piFopPtCYArP242rLTbYbV, programID=A58NQYmJCyDPsc1EfaQZ99piFopPtCYArP242rLTbYbV
	//events from log: <*>{Incremented <*>{F1oMb2TzeNYKwgxJS74nw9z2xF7CL14tiKixoRaxLF79 LST1 150}}
	//parsed 1 instructions of dummy program
	//[<*>{{[11 18 104 9 104 174 59 33] <*>{<*>{11111111111111111111111111111111 0 LST1 50} [<*>{F1oMb2TzeNYKwgxJS74nw9z2xF7CL14tiKixoRaxLF79 true false} <*>{3VPkgde6n22TAD5w69yZbqGJ8ELGdSt7K2kSUjvGYWnR false true}]}}}]
	//incremented LST1 token for 50 amount
}

func Example2() {
	// parsing rpc error
	rpcErr := &jsonrpc.RPCError{}
	err := json.Unmarshal(errJSON, rpcErr)
	if err != nil {
		fmt.Errorf("cannot parse json - %v", err)
		return
	}
	if err, ok := dummy.DecodeCustomError(rpcErr); ok {
		switch {
		case errors.Is(err, dummy.ErrNotImplemented):
			fmt.Printf("not implementeed error: %v", err)
		case errors.Is(err, dummy.ErrInvalidDataFormat):
			fmt.Printf("invalid data format error: %v", err)
		default:
			fmt.Printf("unknown error: %v", err)
		}
	} else {
		fmt.Errorf("cannot decode error - %v", err)
		return
	}

	//output:
	//not implementeed error: NotImplemented(6001): not implemented
}

func Example3() {
	user1, _ := solana.WalletFromPrivateKeyBase58("6Vw4jPBpL6tdAdtQeQ8zTaTV1f8fjda7nBNChswD5cyJ")
	dummy.SetProgramID(dummyProgramID)                                                      // should set this first before find PDA
	userData1, _, _ := dummy.InitializeInstructionUserTokenAmountAccount(user1.PublicKey()) // find PDA

	tx, _ := solana.NewTransaction(
		[]solana.Instruction{
			dummy.NewInitializeInstructionBuilder().
				SetUserAccount(user1.PublicKey()).
				SetUserTokenAmountAccount(userData1).
				//SetSystemProgramAccount(...), // automatically set
				Build(),
			dummy.NewIncrementInstructionBuilder().
				SetData(dummy.UserTokenAmount{
					Token:  "LST1001",
					Amount: 1111,
				}).
				SetUserAccount(user1.PublicKey()).
				SetUserAccount(userData1).
				Build(),
			dummy.NewVersionedMethodInstruction(
				dummy.VersionedStateV1Tuple{
					Elem0: dummy.VersionedStateV1{
						Field1: 1234,
						Field2: "XYZ",
					},
				},
				userData1,
				user1.PublicKey(),
			).Build(),
		},
		solana.Hash{}, // calc recent block hash for real usage
		solana.TransactionPayer(user1.PublicKey()),
	)
	fmt.Printf(tx.MustToBase64())
	//output:
	//AAIAAQMAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFMmuaCG7aeLvkyHijksiDPbp3zkFzKfcTmhaRMJr1UuhsfWeneIkNlSdfsTR9Nf3V6AOHdLxayZkWAsI0KAsLwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMCAwEAAAivr20fDZib7QIBATwLEmgJaK47IQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAcAAABMU1QxMDAxVwQAAAAAAAACAgEAFccl2rBQclanAAAAAAAAAAAAAAAAAA==
}
