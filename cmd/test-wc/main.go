package main

import (
	"fmt"

	"github.com/vultisig/vultisigner/walletcore/core"
)

func main() {
	coinType := core.CoinTypeTHORChain
	name := coinType.ChainID()
	fmt.Print(name)
}
