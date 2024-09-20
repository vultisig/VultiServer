package main

import (
	"fmt"

	"github.com/vultisig/vultisigner/walletcore/core"
)

func main() {
	coinType := core.CoinTypeBitcoin
	name := coinType.GetName()
	fmt.Print(name)
}
