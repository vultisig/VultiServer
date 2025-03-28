# Vultisig Plugin-Verifier-Markeplace

1. Front-end applications that can't integrate with [mobile-tss-lib](https://github.com/vultisig/mobile-tss-lib) will use VultiServer as a TSS server.
2. Fast Vault: Allows creating a 2/2 vault with one mobile device, with VultiServer as the second party. Users can sign transactions with one device.
3. Fast Vault with 2/3: Allows creating a 2/3 vault with two mobile devices and VultiServer as one party. Users can sign transactions with either mobile device without relying on VultiServer to access their crypto assets.

Vultisigner / VultiServer consists of two components:
- API Server: An HTTP server that handles keygen and keysign requests from clients.
- TSS Worker: A service triggered by the API Server to perform the actual TSS operations.

# API Server
## Ping
`/ping` , it provide a simple health check for the Api Server , the return value is `Vultisigner is running`

## Keygen
`POST` `/vault/create`
### Keygen Request
```json
{
  "name": "My Vault",
  "session_id": "session id for key generation",
  "hex_encryption_key": "hex encoded encryption key",
  "hex_chain_code": "hex encoded chain code",
  "local_party_id": "local party id",
  "encryption_password": "password to encryption the generated vault share",
  "email": "email of the user"
}
```
- name: Vault name
- session_id: Key generation session ID (random UUID)
- hex_chain_code: 32-byte hex encoded string
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption
- local_party_id: Identifier for VultiServer in the keygen session
- encryption_password: Password to encrypt the vault share
- email: Email to send the encrypted vault share
- 
### Response

Status Code: OK

## Keysign
`POST` `/vault/sign` , it is used to sign a transaction

### Keysign Request
```json
{
  "public_key": "ECDSA public key of the vault",
  "messages": [
    "hex encoded message 1",
    "hex encoded message 2",
    "hex encoded message N"
  ], 
  "session": "session id for this key sign", 
  "hex_encryption_key": "hex encoded encryption key",
  "derive_path": "derive path for the key sign",
  "is_ecdsa": "is the key sign ECDSA or not",
  "vault_password": "password to decrypt the vault share"
}
```
- public_key: ECDSA public key of the vault
- messages: Hex encoded messages to be signed
- session_id: Key sign session ID (random UUID)
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption
- derive_path: Derive path for the key sign (e.g., BITCOIN: m/44'/0'/0'/0/0)
- is_ecdsa: Boolean indicating if the key sign is for ECDSA
- vault_password: Password to decrypt the vault share

## Get Vault
`GET` `/vault/get/{publicKeyECDSA}` , this endpoint allow user to get the vault information

Note: please set `x-password` header with the password to decrypt the vault share , if the password is empty or incorrect, server will return an error
### Response
```json
{
  "name": "vault name",
  "public_key_ecdsa": "ECDSA public key of the vault",
  "public_key_eddsa": "EdDSA public key of the vault",
  "hex_chain_code": "hex encoded chain code",
  "local_party_id": "local party id"
}
```

## Reshare
`POST` `/vault/reshare` , this endpoint allow user to reshare the vault share

### Reshare Request
```json
{
  "name": "My Vault",
  "public_key": "ECDSA public key of the vault",
  "session_id": "session id for key generation",
  "hex_encryption_key": "hex encoded encryption key",
  "hex_chain_code": "hex encoded chain code",
  "local_party_id": "local party id",
  "old_parties": ["old party id 1", "old party id 2"], 
  "encryption_password": "password to encryption the generated vault share",
  "email": "email of the user",
  "old_reshare_prefix":"old reshare prefix"
}
```
- name: Vault name
- public_key: ECDSA public key
- session_id: Reshare session ID (random UUID)
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption
- hex_chain_code: 32-byte hex encoded string
- local_party_id: Identifier for VultiServer in the reshare session
- old_parties: List of old party IDs
- encryption_password: Password to encrypt the vault share
- email: Email to send the encrypted vault share


## Resend vault share and verification code
`POST` `/vault/resend` , this endpoint allow user to resend the vault share and verification code
Note: user can only request a resend every three minutes

### Resend Request
```json
{
  "public_key_ecdsa": "ECDSA public key of the vault",
  "password": "password to decrypt the vault share",
  "email": "email of the user"
}
```
## Verify code
`GET` `/vault/verify/:public_key_ecdsa/:code` , this endpoint allow user to verify the code
if server return http status code 200, it means the code is valid , other status code means the code is invalid


# Setup Guide

## Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- MinIO client (mc)

## 1. Start Infrastructure Services

First, start the required infrastructure services using Docker Compose:

```
docker compose up -d --remove-orphans
```

## 2. Configure MinIO

Set up MinIO buckets and access:

### Configure MinIO client

```
mc alias set local http://localhost:9000 minioadmin minioadmin
```

### Create required buckets

```
mc mb local/vultiserver
mc mb local/vultiplugin
```

You can verify the buckets were created by visiting the MinIO Console:
- URL: http://localhost:9001
- Username: minioadmin
- Password: minioadmin

## 3. Create Required Directories

### Create directories for vault storage
```
mkdir -p /tmp/vultisigner/server/vaults
mkdir -p /tmp/vultisigner/plugin/vaults
```

### Set appropriate permissions

```
sudo chmod 777 /tmp/vultisigner/server/vaults
sudo chmod 777 /tmp/vultisigner/plugin/vaults
```

## 4. Start Services

Start the services in the following order, each one in a different terminal:

### 4.1. Start Vultisigner Server

```
export VS_CONFIG_NAME=config-server
go run cmd/vultisigner/main.go
```
### 4.2. Start Vultisigner Plugin

```
export VS_CONFIG_NAME=config-plugin
go run cmd/vultisigner/main.go
```

### 4.3. Start Server Worker

```
export VS_CONFIG_NAME=config-server
go run cmd/worker/main.go
```

### 4.4. Start Plugin Worker

```
export VS_CONFIG_NAME=config-plugin
go run cmd/worker/main.go
```

## 5 Run keyGen

On a 5th terminal, run : 

```
curl -X POST http://localhost:8081/vault/create \
-H "Content-Type: application/json" \
-d '{
    "name": "Server2Server-Vault",
    "session_id": "650e8400-e29b-41d4-a716-446655440000",
    "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "hex_chain_code": "2023456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "local_party_id": "2",
    "encryption_password": "your-secure-password",
    "email": "admin2@example.com",
    "start_session": true,
    "parties": ["1", "2"]
}'
```
Then 

```
curl -X POST http://localhost:8080/vault/create \
-H "Content-Type: application/json" \
-d '{
    "name": "Server2Server-Vault",
    "session_id": "650e8400-e29b-41d4-a716-446655440000",
    "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "hex_chain_code": "2023456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "local_party_id": "1",
    "encryption_password": "your-secure-password",
    "email": "admin@example.com",
    "start_session": false
}'
```

## 6 Run keysign 

Once keyGen is done, you can start keysign. You have to replace the ecdsa key by the one appearing in the logs of the keygen.
```
curl -X POST http://localhost:8081/vault/sign \
-H "Content-Type: application/json" \
-d @- << 'EOF'
{
  "public_key": "03b015e2ae364d8f6fff7f2b9fe1760a91e2d41c4a8e91c8750827cea4c3204e5d",
  "messages": ["68656c6c6f"],
  "session": "650e8400-e29b-41d4-a716-446655440000",
  "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "derive_path": "m/44'/0'/0'/0/0",
  "is_ecdsa": true,
  "vault_password": "your-secure-password"
}
EOF
```

Then 

```
curl -X POST http://localhost:8080/vault/sign \
-H "Content-Type: application/json" \
-d @- << 'EOF'
{
  "public_key": "03b015e2ae364d8f6fff7f2b9fe1760a91e2d41c4a8e91c8750827cea4c3204e5d",
  "messages": ["68656c6c6f"],
  "session": "650e8400-e29b-41d4-a716-446655440000",
  "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "derive_path": "m/44'/0'/0'/0/0",
  "is_ecdsa": true,
  "vault_password": "your-secure-password"
}
EOF
```

# Restart everything : 

```
docker compose down
docker compose up -d --remove-orphans
```

Restart all in order : 

```
# Start servers
export VS_CONFIG_NAME=config-server
go run cmd/vultisigner/main.go

export VS_CONFIG_NAME=config-plugin
go run cmd/vultisigner/main.go

# Start workers
export VS_CONFIG_NAME=config-server
go run cmd/worker/main.go

export VS_CONFIG_NAME=config-plugin
go run cmd/worker/main.go
```



## Verification

To verify everything is running correctly:

1. Check Docker containers:

```
docker ps
```

2. Verify MinIO buckets:
```
mc ls local
```






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
