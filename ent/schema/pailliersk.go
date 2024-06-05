package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PaillierSK holds the schema definition for the PaillierSK entity.
type PaillierSK struct {
	ent.Schema
}

// Fields of the PaillierSK.
func (PaillierSK) Fields() []ent.Field {
	return []ent.Field{
		field.String("n").NotEmpty(),
		field.String("lambda_n").NotEmpty(),
		field.String("phi_n").NotEmpty(),
		field.String("p").NotEmpty(),
		field.String("q").NotEmpty(),
	}
}

// Edges of the PaillierSK.
func (PaillierSK) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("vault", Vault.Type).Unique(),
	}
}
