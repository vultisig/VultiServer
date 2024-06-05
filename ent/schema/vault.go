package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Vault holds the schema definition for the Vault entity.
type Vault struct {
	ent.Schema
}

// Fields of the Vault.
func (Vault) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("pub_key").NotEmpty().Match(regexp.MustCompile(`^[0-9a-fA-F]{66}$`)),
		field.JSON("keygen_committee_keys", []string{}), //.NotEmpty(),
		field.String("local_party_key").NotEmpty(),
		field.String("chain_code_hex").NotEmpty().Match(regexp.MustCompile(`^[0-9a-fA-F]{64}$`)),
		field.String("reshare_prefix").Optional(),
	}
}

// Edges of the Vault.
func (Vault) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("paillier_sk", PaillierSK.Type).Ref("vault").Unique(),
		edge.From("ecdsa_local_data", EcdsaLocalData.Type).Ref("vault").Unique(),
		edge.From("eddsa_local_data", EddsaLocalData.Type).Ref("vault").Unique(),
	}
}
