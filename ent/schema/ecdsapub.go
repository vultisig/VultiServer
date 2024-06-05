package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ECDSAPub holds the schema definition for the ECDSAPub entity.
type ECDSAPub struct {
	ent.Schema
}

// Fields of the ECDSAPub.
func (ECDSAPub) Fields() []ent.Field {
	return []ent.Field{
		field.String("curve").NotEmpty().Match(regexp.MustCompile(`^secp256k1$`)),
		field.JSON("coords", []string{}), //.NotEmpty(),
	}
}

// Edges of the EcdsaLocalData.
func (ECDSAPub) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("vault", Vault.Type).Unique(),
	}
}
