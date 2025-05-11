package main

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

var ctx = context.Background()

func main() {
	ethc, err := ethclient.DialContext(ctx, "https://rpc.ankr.com/eth")
	if err != nil {
		panic(err)
	}

}
