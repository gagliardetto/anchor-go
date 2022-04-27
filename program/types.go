package program

import (
    "fmt"
    ag_binary "github.com/gagliardetto/binary"
    "github.com/gagliardetto/solana-go"
)

type IdlProgramAccount struct {
    Authority solana.PublicKey `json:"authority"`
    Data      []byte           `json:"data"`
}

func (obj IdlProgramAccount) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
    err = encoder.Encode(obj.Authority)
    if err != nil {
        return err
    }
    err = encoder.Encode(obj.Data)
    if err != nil {
        return err
    }
    return nil
}

var IdlProgramAccountDiscriminator = [8]byte{24, 70, 98, 191, 58, 144, 123, 158}

func (obj *IdlProgramAccount) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
    {
        discriminator, err := decoder.ReadTypeID()
        if err != nil {
            return err
        }
        if !discriminator.Equal(IdlProgramAccountDiscriminator[:]) {
            return fmt.Errorf("wrong discriminator: wanted %s, got %s", "[24, 70, 98, 191, 58, 144, 123, 158]", fmt.Sprint(discriminator[:]))
        }
    }
    err = decoder.Decode(&obj.Authority)
    if err != nil {
        return err
    }
    err = decoder.Decode(&obj.Data)
    if err != nil {
        return err
    }
    return nil
}
