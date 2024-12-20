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

### 2. Create required directories

Create directories for vault storage with the appropriate permissions

`make gen-vault-dirs`

### 3. Start

Start the services in the following order, each one in a different terminal:
`make plugin-server`
`make signer-server`

Confirm the servers are running via `GET /ping`, it provide a simple health check for the Api Server, the return value is `Vultisigner is running`

Workers:
`make plugin-worker`
`make signer-worker`

For clean restart, do `make down`, and restart the servers/workers again.

### 4. Key Generation

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

## 6. Resharing

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


### configure postgre db : 

```
sudo -u postgres createdb vs-plugins-plugin
sudo -u postgres createdb vs-plugins-vultiserver

```
```
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE \"vs-plugins-plugin\" TO \"Username\";"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE \"vs-plugins-vultiserver\" TO \"Username\";"
```

## Refresh db

```
sudo -u postgres psql
```

### In psql:
```
DROP DATABASE IF EXISTS "vs-plugins-vultiserver";
DROP DATABASE IF EXISTS "vs-plugins-plugin";
```

### Verify they're deleted
```
\l
```


## Create them fresh
```
CREATE DATABASE "vs-plugins-vultiserver";
CREATE DATABASE "vs-plugins-plugin";
```

# use test script 

```
go run scripts/dev/create_vault.go --vault test_1
```

then 
```
go run scripts/dev/create_payroll_policy.go --vault test_1
```

You will be prompted to enter infos, here are dummy ones : 

- Token contract (usdc) `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48`
- recipient address `0x07aE8551Be970cB1cCa11Dd7a11F47Ae82e70E67`
- amount `10`
- `done`
- schedule frequency : `monthly`
