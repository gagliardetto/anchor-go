package program

import (
    "bytes"
    "compress/zlib"
    "context"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/rpc"
    "io/ioutil"
)

func FetchIDL(ctx context.Context, rpc *rpc.Client, address solana.PublicKey) ([]byte, error) {
    var accInfo IdlProgramAccount
    err := rpc.GetAccountDataBorshInto(ctx, address, &accInfo)
    if err != nil {
        return nil, err
    }

    //unCompress
    vv, err := zlib.NewReader(bytes.NewReader(accInfo.Data))
    if err != nil {
        return nil, err
    }
    defer vv.Close()
    bf, err := ioutil.ReadAll(vv)
    return bf, err
}
