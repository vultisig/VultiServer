package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// EddsaLocalData holds the schema definition for the EddsaLocalData entity.
type EddsaLocalData struct {
	ent.Schema
}

// Fields of the EddsaLocalData.
func (EddsaLocalData) Fields() []ent.Field {
	return []ent.Field{
		field.String("xi").Optional(),
		field.String("share_id").Optional(),
		field.String("ks").Optional(),
		field.String("big_xj").Optional(),
		field.String("eddsa_pub").Optional(),
	}
}

// Edges of the EddsaLocalData.
func (EddsaLocalData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("vault", Vault.Type).Unique(),
	}
}
