package chainhelper

import (
	"encoding/hex"
	"errors"
	"fmt"

	v1 "github.com/vultisig/commondata/go/vultisig/keysign/v1"
	"google.golang.org/protobuf/proto"

	"github.com/vultisig/vultisigner/walletcore/core"
	"github.com/vultisig/vultisigner/walletcore/protos/cosmos"
)

type THORChainHelper struct {
	coinType core.CoinType
}

func NewTHORChainHelper() *THORChainHelper {
	return &THORChainHelper{
		coinType: core.CoinTypeTHORChain,
	}
}

var ErrInvalidChain = errors.New("invalid chain")

func (h *THORChainHelper) GetPreSignedInputData(payload *v1.KeysignPayload) ([]byte, error) {
	if payload.Coin.Chain != string(THORChain) {
		return nil, ErrInvalidChain
	}
	thorChainSpecific := payload.GetThorchainSpecific()
	if thorChainSpecific == nil {
		return nil, errors.New("missing thorchain specific")
	}
	if payload.Coin.HexPublicKey == "" {
		return nil, errors.New("missing public key")
	}
	publicKeyBytes, err := hex.DecodeString(payload.Coin.HexPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	fromAddrData := core.GetAnyAddressData(payload.Coin.Address, h.coinType)
	var messages []*cosmos.Message
	if thorChainSpecific.IsDeposit {
		coin := &cosmos.THORChainCoin{
			Asset: &cosmos.THORChainAsset{
				Chain:  "THOR",
				Symbol: "RUNE",
				Ticker: "RUNE",
				Synth:  false,
			},
			Amount:   payload.ToAmount,
			Decimals: 8,
		}
		messages = []*cosmos.Message{
			{
				MessageOneof: &cosmos.Message_ThorchainDepositMessage{
					ThorchainDepositMessage: &cosmos.Message_THORChainDeposit{
						Signer: fromAddrData,
						Memo:   payload.GetMemo(),
						Coins: []*cosmos.THORChainCoin{
							coin,
						},
					},
				},
			},
		}
	} else {
		toAddrData := core.GetAnyAddressData(payload.ToAddress, h.coinType)
		messages = []*cosmos.Message{
			{
				MessageOneof: &cosmos.Message_ThorchainSendMessage{
					ThorchainSendMessage: &cosmos.Message_THORChainSend{
						FromAddress: fromAddrData,
						Amounts: []*cosmos.Amount{
							{
								Denom:  "rune",
								Amount: payload.ToAmount,
							},
						},
						ToAddress: toAddrData,
					},
				},
			},
		}
	}
	input := cosmos.SigningInput{
		SigningMode:   cosmos.SigningMode_Protobuf,
		AccountNumber: thorChainSpecific.AccountNumber,
		ChainId:       h.coinType.ChainID(),
		Fee: &cosmos.Fee{
			Gas: 20000000,
		},
		Sequence:  thorChainSpecific.Sequence,
		Messages:  messages,
		Mode:      cosmos.BroadcastMode_SYNC,
		PublicKey: publicKeyBytes,
	}
	if payload.Memo != nil {
		input.Memo = *payload.Memo
	}
	return proto.Marshal(&input)
}
