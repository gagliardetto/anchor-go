package program

import (
    "context"
    "fmt"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
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
    fmt.Println(string(idlBytes), err)
    fmt.Println(err, pid.String())
}
