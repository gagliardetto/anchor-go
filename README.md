# anchor-go

![logo](logo.png)

## usage

```
# Install anchor-go
go install github.com/gagliardetto/anchor-go@latest

# Generate code from an IDL file
anchor-go --idl /path/to/idl.json --output ./generated --program-id 0123456789abcdef0123456789abcdef0123456789
```

## Features

- [x] instructions
- [x] accounts
- [x] events
- [x] types
- [x] handle tuple types
- [x] constants
- [ ] error parsing


## what is anchor-go?

`anchor-go` generates Go clients for [Solana](https://solana.com/) programs (smart contracts) written using the [anchor](https://github.com/solana-foundation/anchor) framework.

This version of `anchor-go` only supports the IDL format of anchor starting with version 0.30.0.

If you have an older version of anchor, you can use `anchor idl convert <my-old-idl.json>` to convert it to the new format.

