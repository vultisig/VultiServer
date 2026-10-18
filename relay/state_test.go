package relay

import (
	"testing"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

func keyImportVault() *vaultType.Vault {
	return &vaultType.Vault{
		PublicKeyEcdsa: "root-ecdsa",
		PublicKeyEddsa: "root-eddsa",
		KeyShares: []*vaultType.Vault_KeyShare{
			{PublicKey: "root-ecdsa", Keyshare: "ks-root-ecdsa"},
			{PublicKey: "root-eddsa", Keyshare: "ks-root-eddsa"},
			{PublicKey: "evm", Keyshare: "ks-arbitrum"},
			{PublicKey: "btc", Keyshare: "ks-bitcoin"},
			{PublicKey: "evm", Keyshare: "ks-bsc"},
			{PublicKey: "evm", Keyshare: "ks-ethereum"},
		},
		ChainPublicKeys: []*vaultType.Vault_ChainPublicKey{
			{Chain: "Arbitrum", PublicKey: "evm"},
			{Chain: "Bitcoin", PublicKey: "btc"},
			{Chain: "BSC", PublicKey: "evm"},
			{Chain: "Ethereum", PublicKey: "evm"},
		},
	}
}

func TestGetChainLocalStateSharedPublicKey(t *testing.T) {
	l := &LocalStateAccessorImp{Vault: keyImportVault()}
	for chain, want := range map[string]string{
		"Arbitrum": "ks-arbitrum",
		"Bitcoin":  "ks-bitcoin",
		"BSC":      "ks-bsc",
		"ethereum": "ks-ethereum",
	} {
		pk, ks, err := l.GetChainLocalState(chain)
		if err != nil {
			t.Fatalf("%s: %v", chain, err)
		}
		if ks != want {
			t.Errorf("%s: got keyshare %q, want %q", chain, ks, want)
		}
		if pk == "" {
			t.Errorf("%s: empty public key", chain)
		}
	}
}

func TestGetChainLocalStateChainKeyEqualsRootKey(t *testing.T) {
	// a raw secp256k1 key imported as both the root and the Ethereum chain key
	l := &LocalStateAccessorImp{Vault: &vaultType.Vault{
		PublicKeyEcdsa: "k",
		PublicKeyEddsa: "root-eddsa",
		KeyShares: []*vaultType.Vault_KeyShare{
			{PublicKey: "k", Keyshare: "ks-root"},
			{PublicKey: "root-eddsa", Keyshare: "ks-root-eddsa"},
			{PublicKey: "k", Keyshare: "ks-ethereum"},
		},
		ChainPublicKeys: []*vaultType.Vault_ChainPublicKey{{Chain: "Ethereum", PublicKey: "k"}},
	}}
	_, ks, err := l.GetChainLocalState("Ethereum")
	if err != nil || ks != "ks-ethereum" {
		t.Fatalf("got %q, %v; want ks-ethereum", ks, err)
	}
}

func TestGetChainLocalStateOneSharePerKey(t *testing.T) {
	// vault that stored a single keyshare for a key used by several chains
	l := &LocalStateAccessorImp{Vault: &vaultType.Vault{
		KeyShares: []*vaultType.Vault_KeyShare{{PublicKey: "evm", Keyshare: "ks-evm"}},
		ChainPublicKeys: []*vaultType.Vault_ChainPublicKey{
			{Chain: "Ethereum", PublicKey: "evm"},
			{Chain: "BSC", PublicKey: "evm"},
		},
	}}
	_, ks, err := l.GetChainLocalState("BSC")
	if err != nil || ks != "ks-evm" {
		t.Fatalf("got %q, %v; want ks-evm", ks, err)
	}
}

func TestGetChainLocalStateUnknownChain(t *testing.T) {
	l := &LocalStateAccessorImp{Vault: keyImportVault()}
	if _, _, err := l.GetChainLocalState("Solana"); err == nil {
		t.Fatal("expected error for a chain the vault does not hold")
	}
}
