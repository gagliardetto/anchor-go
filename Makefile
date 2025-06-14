restaking:
	go build && rm -rf ./generated/restaking && \
	./solana-anchor-go -src=./example/restaking_idl.json -pkg=restaking -dst=./generated/restaking && \
	go test ./generated/restaking && \
	go test ./example/restaking_test.go

test:
	make dummy && make restaking