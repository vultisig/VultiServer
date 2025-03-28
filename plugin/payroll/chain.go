package payroll

//do a mapping : according to the chain ID, it should return isecdsa or iseddsa

func IsECDSA(chainID int64) bool {
	return chainID == 1 || chainID == 5 || chainID == 10 || chainID == 137 || chainID == 42161
}

// do a mapping : according to the chain ID, it should return the chain name
func GetChainName(chainID int64) string {
	switch chainID {
	case 1:
		return "Ethereum"
	case 5:
		return "Goerli"
	case 10:
		return "Optimism"
	case 137:
		return "Polygon"
	case 42161:
		return "Arbitrum"
	}
	return ""
}
