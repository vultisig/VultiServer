package chainhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/vultisig/commondata/go/vultisig/keysign/v1"
)

func TestTHORChainDerivationPath(t *testing.T) {
	helper := NewTHORChainHelper()
	result := helper.coinType.DerivationPath()
	assert.Equal(t, result, "m/44'/931'/0'/0/0")
}
func getTestKeysignPayload() v1.KeysignPayload {
	return v1.KeysignPayload{
		Coin: &v1.Coin{
			Chain:         "THORChain",
			Ticker:        "RUNE",
			Decimals:      8,
			Address:       "thor10stcxwypezd4pqwsdymu2p9hq90wtau6j4uljg",
			IsNativeToken: true,
			HexPublicKey:  "02bd71faf6447dd28ecc7936729c543e8de0483c9641ed65fcd4f223b010263c67",
		},
		ToAddress: "thor1vzltn37rqccwk95tny657au9j2z072dhgstcmn",
		ToAmount:  "1000000",
		BlockchainSpecific: &v1.KeysignPayload_ThorchainSpecific{
			ThorchainSpecific: &v1.THORChainSpecific{
				AccountNumber: 1,
				Sequence:      0,
				Fee:           0,
				IsDeposit:     false,
			},
		},

		VaultPublicKeyEcdsa: "020503826804dcf347bb5c98331f10ad388fdbc935adf775154089acd89f2ce9dd",
		VaultLocalPartyId:   "Server-1234",
	}
}
func TestTHORChainHelper_GetPreSignedInputData(t *testing.T) {
	helper := NewTHORChainHelper()
	payload := getTestKeysignPayload()
	result, err := helper.GetPreSignedImageHash(&payload)
	if err != nil {
		t.Errorf("GetPreSignedInputData() error = %v", err)
	}
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0], "cac1691056905f68cec68ac322b4a067c511030996c876bd47d52fdbab34dd4a")
}
