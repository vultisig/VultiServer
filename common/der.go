package common

import (
	"encoding/asn1"
	"math/big"
)

type ECDSASignature struct {
	R, S *big.Int
}

func GetDerSignature(r, s []byte) ([]byte, error) {
	rInt := new(big.Int).SetBytes(r)
	sInt := new(big.Int).SetBytes(s)
	sig := ECDSASignature{R: rInt, S: sInt}
	der, err := asn1.Marshal(sig)
	if err != nil {
		return nil, err
	}
	return der, nil
}
