package mapper

import (
	"strings"
	"vultisigner/internal/models"
	"vultisigner/internal/types"
)

type KeyGenerationMapper struct{}

func (s KeyGenerationMapper) ToModel(req types.KeyGeneration) models.KeyGeneration {
	return models.KeyGeneration{
		Key:       req.Key,
		Parties:   strings.Join(req.Parties, ","),
		Session:   req.Session,
		ChainCode: req.ChainCode,
	}
}

func (s KeyGenerationMapper) ToAPI(model models.KeyGeneration) types.KeyGeneration {
	return types.KeyGeneration{
		Key:       model.Key,
		Parties:   strings.Split(model.Parties, ","),
		Session:   model.Session,
		ChainCode: model.ChainCode,
	}
}
