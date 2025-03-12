# Vulti-Plugin-Server

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

Verify the buckets were created by visiting the MinIO Console: http://localhost:9001 (username: minioadmin, password: minioadmin)


### 2. Start the servers and workers

Start the services in the following order, each one in a different terminal:

Servers:
```sh
make verifier-server
make plugin-server
```

Workers:
```sh
make verifier-worker
make plugin-worker
```

For clean restart, do `make down`, and restart the servers/workers again.

### 3. Key Generation

- name: Vault name
- session_id: Key generation session ID (random UUID)
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption 
- hex_chain_code: 32-byte hex encoded string
- local_party_id: Identifier for VultiServer in the keygen session
- encryption_password: Password to encrypt the generated vault share
- email: Email to send the encrypted vault share
- start_session: Flag that specifies if a new session should be initiated
- parties: List of expcted parties to participate

```sh
curl -X POST http://localhost:8081/vault/create \
-H "Content-Type: application/json" \
-d '{
    "name": "Server-2-Server-Vault",
    "session_id": "913e7402-e29b-41d4-a716-446655440000",
    "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "hex_chain_code": "2023456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "local_party_id": "1",
    "encryption_password": "your-secure-password",
    "email": "admin1@example.com",
    "start_session": true,
    "parties": ["1", "2"]
}'
```

Then

```sh
curl -X POST http://localhost:8080/vault/create \
-H "Content-Type: application/json" \
-d '{
    "name": "Server-2-Server-Vault",
    "session_id": "913e7402-e29b-41d4-a716-446655440000",
    "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "hex_chain_code": "2023456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "local_party_id": "2",
    "encryption_password": "your-secure-password",
    "email": "admin2@example.com",
    "start_session": false
}'
```

### 4. Key Signing

Before starting the keysign, make sure to replace the public key by the one appearing in the logs of the keygen. 

- public_key: ECDSA public key of the vault
- messages: List of hex encoded messages to be signed
- session_id: Key sign session ID (random UUID)
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption key
- derive_path: Derive path for the key sign (e.g., BITCOIN: m/44'/0'/0'/0/0)
- is_ecdsa: Boolean indicating if the key sign is for ECDSA
- vault_password: Password to decrypt the vault share

```sh
curl -X POST http://localhost:8081/vault/sign \
-H "Content-Type: application/json" \
-d '{
  "public_key": "0259091f366a9664c020b4d97b1c5d87573a5c4b223744501f23a61c8fe394e5d3",
  "messages": ["68656c6c6f"],
  "session": "877e6502-e29b-41d4-a716-446655440000",
  "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "derive_path": "m/44'/0'/0'/0/0",
  "is_ecdsa": true,
  "vault_password": "your-secure-password",
  "start_session": true,
  "parties": ["1", "2"]
}'
```

Then

```sh
curl -X POST http://localhost:8080/vault/sign \
-H "Content-Type: application/json" \
-d '{
  "public_key": "0259091f366a9664c020b4d97b1c5d87573a5c4b223744501f23a61c8fe394e5d3",
  "messages": ["68656c6c6f"],
  "session": "877e6502-e29b-41d4-a716-446655440000",
  "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "derive_path": "m/44'/0'/0'/0/0",
  "is_ecdsa": true,
  "vault_password": "your-secure-password",
  "start_session": false,
  "parties": ["1", "2"]
}'
```

## 5. Resharing

Allow user to reshare the vault share

- name: Vault name
- public_key: ECDSA public key
- session_id: Reshare session ID (random UUID)
- hex_encryption_key: 32-byte hex encoded string for encryption/decryption
- hex_chain_code: 32-byte hex encoded string
- local_party_id: Identifier for VultiServer in the reshare session
- old_parties: List of old party IDs
- encryption_password: Password to encrypt the vault share
- email: Email to send the encrypted vault share
- old_reshare_prefix: ...

```sh
curl -X POST http://localhost:8081/vault/reshare \
-H "Content-Type: application/json" \
-d '{
    "name": "Server2Server-Vault",
    "public_key": ",
    "session_id": "650e8400-e29b-41d4-a716-446655440000",
    "hex_encryption_key": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "hex_chain_code": "2023456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "local_party_id": "1",
    "old_parties": ["1", "2"], 
    "encryption_password": "your-secure-password",
    "email": "admin@example.com",
    "start_session": false,
    "old_reshare_prefix": "old reshare prefix"
}'
```

<!-- ## 6. Verify code

`GET` `/vault/verify/:public_key_ecdsa/:code` , this endpoint allow user to verify the code
if server return http status code 200, it means the code is valid , other status code means the code is invalid -->

# 7. Test DCA Plugin Execution

1. Create 2-of-2 Fast vault on some device (e.g. computer)
2. Backup the vault shares (device share, vultiserver share)
3. Import the vault via QR code into Vulticonnect
4. Import the vault into another device (e.g. phone)
5. Import each backup into the relevant local S3 folder (vultiplugin, vultiserver) named as `publickey`.bak.vult (for local testing)
6. Change local participant ids (`PluginPartyID, VerifierPartyID`) to match those from the backups (for local testing, will be removed once the new reshare process is implemented) 
7. Change the hardcoded public key (will be removed once it is properly derived)
8. Change the hardcoded vault password (will be removed)
9. Create a DCA policy through the UI

```sh
export VAULT_ADDRESS=0x5582df2D22194AF8201997D750e80fd8140387c2
export RPC_URL=http://127.0.0.1:8545
export PRIVATE_KEY=ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 
export TOKEN_ADDRESS=0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
```

10. Send some amount of native ETH to the vault address
```sh
  cast send $VAULT_ADDRESS --value 10ether --rpc-url $RPC_URL --private-key $PRIVATE_KEY
  cast balance $VAULT_ADDRESS  --rpc-url $RPC_URL
  cast nonce $VAULT_ADDRESS  --rpc-url $RPC_URL
```

11. Mint some amount of ERC20 Token. If "-token" is not present, the script will default to minting WETH.
```sh
  go run scripts/dev/mint_erc20/main.go -vault-address $VAULT_ADDRESS -token $TOKEN_ADDRESS
```

## Create Payroll plugin policy

```sh
  go run scripts/dev/create_payroll_policy/main.go -state-dir $STATE_DIR -vault $VAULT_NAME
```

- Token contract (usdc) `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48`
- Polygon token contract `0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174`
Enter recipients and amounts one by one - enter 'done' when finished
- Enter a recipient address: `0x07aE8551Be970cB1cCa11Dd7a11F47Ae82e70E67`
- chain id : `137`
- Enter the amount for this recipient: `100`
- `done`
- schedule frequency : `5-minutely`

```sh
  go run scripts/dev/create_payroll_policy/main.go -state-dir $STATE_DIR -vault $VAULT_NAME
```

- Enter token contract: `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` UCDC
Enter recipients and amounts one by one - enter 'done' when finished
- Enter a recipient address: `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266`
- Enter the amount for this recipient: `1000`
- `done`
- schedule frequency : `5-minutely`