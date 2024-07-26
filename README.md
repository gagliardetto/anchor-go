# anchor-go

![logo](logo.png)

## usage

```bash
anchor-go --src=/path/to/idl.json
```

Generated Code will be generated and saved to `./generated/`.

## TODO

- [x] instructions
- [x] accounts
- [x] types
- [ ] events
- [ ] errors
- [ ] handle tuple types
- [ ] constants

## Future Development

1. **Fork and Maintain Anchor-Go Repository**: We have noted that an anchor-go repository exists in a deprecated state; this shall be our starting point. We shall fork, update, and maintain the repository to align it with contemporary standards and requirements. This repository will be modified to better serve not only our needs but also those of the wider Solana ecosystem.
2. **Improve On-Chain and Off-Chain Interaction**: We will enable more seamless on-chain to off-chain data synchronization, on-chain event subscription management, and network status monitoring services through the utilization of Go's concurrency features and the performance advantages that the language inherently brings.
3. **IDL Support in GoLang**: Will provide full IDL support to build the SDK in GoLang, with the help of the Anchor IDL specification; thus, it's easily adopted and becomes integrated smoothly into the GoLang development environment.
4. **Comprehensive Documentation and Community Support**: We'll have up-to-date documentation that will also come with lifetime support to developers in the community, including detailed guides, lots of examples, and a maintained team, ready to respond to issues and implement feedback
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
git clone https://github.com/metaplex-foundation/metaplex-program-library.git
cd metaplex-program-library
anchor idl parse -f candy-machine/program/src/lib.rs -o nft_candy_machine_v2.json
anchor-go --src=nft_candy_machine_v2.json
```

Note
----

- anchor-go is in active development, so all APIs are subject to change.
- This code is unaudited. Use at your own risk.
