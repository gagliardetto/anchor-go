package program

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFetchIDL(t *testing.T) {
	client := rpc.New(rpc.DevNet_RPC)
	pid, _ := solana.PublicKeyFromBase58("cndy3Z4yapfJBmL3ShUp5exZKqR3z33thTzeNMm2gRZ")
	pid, _ = IDLAddress(pid)
	idlBytes, err := FetchIDL(context.TODO(), client, pid)
	if err != nil {
		panic(err)
	}
	require.Equal(t, pid.String(), "CggtNXgCye2qk7fLohonNftqaKT35GkuZJwHrRghEvSF")
	got := base58.Encode(idlBytes)
	require.Equal(t, len(got), 12605)
}
