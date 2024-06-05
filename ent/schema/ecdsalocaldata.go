package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// EcdsaLocalData holds the schema definition for the EcdsaLocalData entity.
type EcdsaLocalData struct {
	ent.Schema
}

// Fields of the EcdsaLocalData.
func (EcdsaLocalData) Fields() []ent.Field {
	return []ent.Field{
		field.String("n_tilde_i").NotEmpty(),
		field.String("h1i").NotEmpty(),
		field.String("h2i").NotEmpty(),
		field.String("alpha").NotEmpty(),
		field.String("beta").NotEmpty(),
		field.String("p").NotEmpty(),
		field.String("q").NotEmpty(),
		field.String("xi").NotEmpty(),
		field.String("share_id").NotEmpty(),
		field.JSON("ks", []string{}),           //.NotEmpty(),
		field.JSON("n_tilde_j", []string{}),    //.NotEmpty(),
		field.JSON("h1j", []string{}),          //.NotEmpty(),
		field.JSON("h2j", []string{}),          //.NotEmpty(),
		field.JSON("big_xj", []ECDSAPub{}),     //.NotEmpty(),
		field.JSON("paillier_pks", []string{}), //.NotEmpty(),
	}
}

// Edges of the EcdsaLocalData.
func (EcdsaLocalData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("vault", Vault.Type).Unique(),
	}
}
