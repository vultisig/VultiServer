# Vultisig Plugin-Verifier-Markeplace

## Dev setup

### Prerequisites

- Git
- Go >= 1.21
- Docker, Docker Compose


### 1. Start the required infrastructure

- Postgres
- Redis
- MinIO (for faster development, buckets are auto created and configured during this step via `create_buckets.sh`)

`make up`

Verify the buckets were created by visiting the MinIO Console: 
http://localhost:9001 (username: minioadmin, password: minioadmin)


### 2. Start the Verifier, Plugin and Marketplace

Start each service in a different tab:

Verifer

```sh
make verifier-server
make verifier-worker
```

Plugin

```sh
make plugin-server
make plugin-worker
```

Marketplace

```sh
  cd plugins-ui
  npm ci
  npm run dev
```

For a clean restart, run `make down` and restart the servers and workers (consider deleting the volumes and images for a fresh start)


### 3. Add new plugin listing and pricing in the Marketplace

**Admin User**

Create an admin user

```sh
  go run ./scripts/dev/create_verifier_admin/main.go -username=admin -password=supersecret
```

Log in to get an auth token, for subsequent requests: (`myauthtoken`)

```sh
curl --location localhost:8080/login --request POST \
--header 'Content-Type: application/json' \
--data '{"username":"admin","password":"supersecret"}'
```

**Pricing**

Create a pricing plan

```sh
curl --location localhost:8080/pricings --request POST \
--header 'Authorization: Bearer myauthtoken' \
--header 'Content-Type: application/json' \
--data '{
  "type": "FREE",
  "amount": 0,
  "metric": "FIXED"
}'
```

**Plugin Listing**

Add new plugin listing

```sh
curl --location localhost:8080/plugins --request POST \
--header 'Authorization: Bearer myauthtoken' \
--header 'Content-Type: application/json' \
--data '{
    "title": "DCA Plugin",
    "type": "dca",
    "description": "Dollar cost averaging plugin automation",
    "metadata": "{\"foo\": \"bar\"}",
    "server_endpoint": "http://localhost:8081",
    "pricing_id": "12345678-abcd-1234-5678-123456789abc"
}'
```


### 4. Test the DCA Plugin execution (with Vault from production imported into Vulticonnect)

1. Create a 2-of-3 vault, allowing you to sign plugin policies using only two user devices
2. Back up both device shares by exporting them from the Vultisig apps on each device
3. Import the vault via QR code into VaultConnect, allowing you to sign plugin policies

4. There is an easier way to test with vaults created via script, however, it is not easy to import these vaults into Vulticonnect and thus are not suitable for full end-to-end testing. And since we are instead testing with vaults from production, there are additional steps required to bring in the missing production data locally.
  - Rename each backup as `<PUBLIC_KEY>.bak` and import each into the corresponding local S3 folder (`vultisig-plugin`, `vultisig-verifier`)

5. The new MPC resharing is not yet merged, and as of now, there is no way to install a plugin that would bring the Verifier and Plugin as part of a user vault. This would allow the Plugin and Verifier to meet the vault threshold and sign transactions. For local testing, we are using two shares from the user vault: first one for the Verifier and the second one for the Plugin. As a result, there are some hardcoded values that should match the vault being used in Vulticonnect.
  - Change the `PluginPartyID` and `VerifierPartyID` (in common/util.go) to match those from the Vultisig app -> Vault settings -> Details.
  - Change the hardcoded `vaultPassword` to match the vault password and the `hexEncryptionKey` to match the `hexChainCode` that we can get if we execute `await window.vultisig.getVaults()`

6. Create a DCA policy through the UI
7. The policy execution will start, so it is essential to ensure that the vault address has sufficient balance for the token and amount specified in the policy.

```sh
export RPC_URL=http://127.0.0.1:8545 # from the local ethereum fork
export PRIVATE_KEY=ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 # from the local ethereum fork
export VAULT_ADDRESS=0x5582df2D22194AF8201997D750e80fd8140387c2 # from the vultisig app
```

Send some amount of ETH to the vault address

```sh
  cast send $VAULT_ADDRESS --value 10ether --rpc-url $RPC_URL --private-key $PRIVATE_KEY
```

Mint some amount of the ERC20 token used as a source in the policy. If "-token" is not present, the script will default to minting WETH.

```sh
  export TOKEN_ADDRESS=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 # USDC in this case
  go run scripts/dev/mint_erc20/main.go -vault-address $VAULT_ADDRESS -token $TOKEN_ADDRESS
```
