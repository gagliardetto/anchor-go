package example

import (
	_ "embed"
	"fmt"
	"github.com/encrypt-x/solana-anchor-go/generated/restaking"
	"github.com/gagliardetto/solana-go"
)

var restakingProgramID = solana.MustPublicKeyFromBase58("fragfP1Z2DXiXNuDYaaCnbGvusMP1DNQswAqTwMuY6e")

func Example4() {
	user1, _ := solana.WalletFromPrivateKeyBase58("6Vw4jPBpL6tdAdtQeQ8zTaTV1f8fjda7nBNChswD5cyJ")
	fragSOLMintAddress, _ := solana.PublicKeyFromBase58("FRAGsJAbW4cHk2DYhtAWohV6MUMauJHCFtT1vGvRwnXN")

	restaking.SetProgramID(restakingProgramID)                                                       // should set this first before find PDA
	fragSOLFundAddress, _, _ := restaking.FundDepositTokenInstructionFundAccount(fragSOLMintAddress) // find PDA

	tx, _ := solana.NewTransaction(
		[]solana.Instruction{
			restaking.NewFundUpdateDefaultProtocolFeeRateInstructionBuilder().
				SetFundAccount(fragSOLFundAddress).
				SetRequest(restaking.FundUpdateDefaultProtocolFeeRateRequestV1Tuple{
					Elem0: restaking.FundUpdateDefaultProtocolFeeRateRequestV1{
						DefaultProtocolFeeRate: 10,
					},
				}).
				// SetAdminAccount(...). // automatically set
				// SetSystemProgramAccount(...). // automatically set
				Build(),
			restaking.NewLogMessageInstruction("traceparent: 00-0af7651916cd43dd8448eb211c80319c-b9c7c989f97918e1-01\ntracestate: congo=ucfJifl5GOE,rojo=00f067aa0ba902b7").
				Build(),
		},
		solana.Hash{}, // calc recent block hash for real usage
		solana.TransactionPayer(user1.PublicKey()),
	)
	//fmt.Println(tx.String())
	fmt.Printf(tx.MustToBase64())
	//output:
	//AAEAAgR3HicvsoIYj5uf0anX5ZursGygzmpiQ3vOoF5XJFuC7SJLWwyDI7BtyZou/DbjBDKJSfl66jNt2wx0hNUXqeBHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJ9CHDJI615fn5u1m6o0UWxEI9cGhKuWamLZSL/7HtQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgMDAAECC89G0OK1pwykAAAAAwCEAZQELCLKBVNzeAAAAHRyYWNlcGFyZW50OiAwMC0wYWY3NjUxOTE2Y2Q0M2RkODQ0OGViMjExYzgwMzE5Yy1iOWM3Yzk4OWY5NzkxOGUxLTAxCnRyYWNlc3RhdGU6IGNvbmdvPXVjZkppZmw1R09FLHJvam89MDBmMDY3YWEwYmE5MDJiNw==
}
