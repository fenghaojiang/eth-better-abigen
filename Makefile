.PHONY: build 

build: 
	go build -o build/abigen ./cmd/abigen

gen:
	go generate ./example/erc20/...