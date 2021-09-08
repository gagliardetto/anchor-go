## usage

```bash
anchor-go --src=/path/to/idl.json
```

Generated Code will be generated and saved to `./generated/`.

## TODO

- [x] instructions
- [x] accounts
- [x] types
- [ ] state
- [ ] events
- [ ] errors
- [ ] handle tuple types


## what is anchor-go?

`anchor-go` generates Go clients for [Solana](https://solana.com/) programs (smart contracts) written using the [anchor](https://github.com/project-serum/anchor) framework.

## what is anchor?

Link: https://github.com/project-serum/anchor

```
Anchor is a framework for Solana's Sealevel runtime providing several convenient developer tools for writing smart contracts.
```

## I have an anchor program; how do I generate a Go client for it? (step by step)

### example 1: metaplex nft candy machine

```bash
git clone https://github.com/metaplex-foundation/metaplex.git
cd metaplex
anchor idl parse -f rust/nft-candy-machine/src/lib.rs -o nft_candy_machine.json
anchor-go --src=nft_candy_machine.json
```
