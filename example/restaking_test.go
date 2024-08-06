package example

import (
	_ "embed"
	"fmt"
	"github.com/encrypt-x/solana-anchor-go/generated/restaking"
	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

var restakingProgramID = solana.MustPublicKeyFromBase58("fragfP1Z2DXiXNuDYaaCnbGvusMP1DNQswAqTwMuY6e")

func Example4() {
	user1, _ := solana.WalletFromPrivateKeyBase58("6Vw4jPBpL6tdAdtQeQ8zTaTV1f8fjda7nBNChswD5cyJ")
	fragSOLMintAddress, _ := solana.PublicKeyFromBase58("FRAGsJAbW4cHk2DYhtAWohV6MUMauJHCFtT1vGvRwnXN")

	restaking.SetProgramID(restakingProgramID)                                                       // should set this first before find PDA
	fragSOLFundAddress, _, _ := (*restaking.FundInitialize)(nil).FindFundAddress(fragSOLMintAddress) // find PDA for instruction, it is safe for nil receiver

	tx, _ := solana.NewTransaction(
		[]solana.Instruction{
			restaking.NewFundAddWhitelistedTokenInstructionBuilder().
				SetFundAccount(fragSOLFundAddress).
				SetRequest(restaking.FundAddWhitelistedTokenRequestV1Tuple{
					Elem0: restaking.FundAddWhitelistedTokenRequestV1{
						Token: solana.NewWallet().PublicKey(),
						TokenCap: ag_binary.Uint128{
							Hi: 10,
							Lo: 10,
						},
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
	//AAEAAgR3HicvsoIYj5uf0anX5ZursGygzmpiQ3vOoF5XJFuC7SJLWwyDI7BtyZou/DbjBDKJSfl66jNt2wx0hNUXqeBHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJ9CHDJI615fn5u1m6o0UWxEI9cGhKuWamLZSL/7HtQwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgMDAAECOd9+//W3vPEXAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMAhAGUBCwiygVTc3gAAAB0cmFjZXBhcmVudDogMDAtMGFmNzY1MTkxNmNkNDNkZDg0NDhlYjIxMWM4MDMxOWMtYjljN2M5ODlmOTc5MThlMS0wMQp0cmFjZXN0YXRlOiBjb25nbz11Y2ZKaWZsNUdPRSxyb2pvPTAwZjA2N2FhMGJhOTAyYjc=
}
