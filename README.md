# Vultisigner

## cmd/vultisigner

### Vault Management

#### [POST] /vault

Create a new vault 

**Request:**
```json
{
  "name": "My Vault",
  "encryption_password": "hex encoded password for message encryption"
}
```

**Response (initial):**
```json
{
  "name": "My Vault", 
  "session_id": "",
  "hex_encryption_key": "",
  "hexChainCode": "80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0"
}
```

#### [POST] /vault/upload


#### [GET] /vault/download/{publicKeyECDSA}

## cmd/worker

This component handles the background jobs for key generation and transaction signing using Asynq.

### Key Generation

#### Job: Key Generation

- Triggered automatically upon vault creation.
- Joins the key generation session using the provided session ID.
- Stores the key generation data securely upon completion.


### Transaction Signing

#### Job: Transaction Signing

- Triggered when a transaction signing request is made.
- Joins the transaction signing session using the provided session ID.
- Waits for all parties to join and complete the signing process.
