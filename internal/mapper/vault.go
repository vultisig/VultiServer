package mapper

import (
	"strings"
	"vultisigner/internal/models"
	"vultisigner/internal/types"
)

type VaultMapper struct{}

func (s VaultMapper) ToModel(req types.Vault) models.Vault {
	return models.Vault{
		PubKey:    req.Key,
		Parties:   strings.Join(req.Parties, ","),
		Session:   req.Session,
		ChainCode: req.ChainCode,
	}
}

func (s VaultMapper) ToAPI(model models.Vault) types.Vault {
	return types.Vault{
		Key:       model.Key,
		Parties:   strings.Split(model.Parties, ","),
		Session:   model.Session,
		ChainCode: model.ChainCode,
	}
}
