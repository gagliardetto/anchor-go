# anchor-go

![logo](logo.png)

## Usage

```bash
$ go build
$ ./solana-anchor-go -src=./idl/fragmetric/dummy.json -pkg=dummy -dst=./generated/dummy
```

Generated Code will be generated and saved to `./generated/`.
And check `./example/dummy_test.go` for generated code usage.

## TODO
- [x] instructions
- [x] accounts
- [x] types
- [x] events
- [x] errors
- [ ] handle tuple types
- [ ] constants (?)
