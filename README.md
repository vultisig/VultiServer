# Vultisigner

## cmd/vultisigner

### Vault Management

#### [POST] /vaults

Create a new vault with specified parties. Once the vault is created, the key generation process will start, spinning up a background job (using `cmd/worker` with `asynq`) to join the TSS-coordinator keygen session. During this process, the keygen status will be `pending`, and upon completion, it will change to `completed` with public keys available.

**Request:**
\```json
{
  "name": "My Vault",
  "description": "Description of the vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ]
}
\```

**Response (initial):**
\```json
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "description": "Description of the vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "keygen_status": "pending",
  "created_at": "timestamp"
}
\```

**Response (after keygen completion):**
\```json
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "description": "Description of the vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "keygen_status": "completed",
  "shares": {
    "ecdsa": [
      {"party_id": "party_1_id", "public_key": "public_key_data_1"},
      {"party_id": "party_2_id", "public_key": "public_key_data_2"},
      {"party_id": "Vultisigner", "public_key": "public_key_data_vultisigner"}
    ],
    "eddsa": [
      {"party_id": "party_1_id", "public_key": "public_key_data_1"},
      {"party_id": "party_2_id", "public_key": "public_key_data_2"},
      {"party_id": "Vultisigner", "public_key": "public_key_data_vultisigner"}
    ]
  },
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
\```

#### [GET] /vaults/{vault_id}/keygen/status

View the status of the key generation job for a specific vault.

**Response:**
\```json
{
  "vault_id": "unique_vault_id",
  "keygen_status": "pending", // or "in_progress", "completed", "failed"
  "session_id": "unique_session_id",
  "updated_at": "timestamp"
}
\```

### Transaction Policy Management

#### [POST] /vaults/{vault_id}/transaction_policies/request_update

Request an update to the transaction policy for a specific vault. This generates data that other parties must sign to approve the update. They should sign the contents of `new_policy` and send the signatures to the `/vaults/{vault_id}/transaction_policies/{update_request_id}/approve` endpoint for it to take effect.

**Request:**
\```json
{
  "new_policy": {
    "policy_id": "unique_policy_id",
    "rules": [
      {"condition": "wait_time", "value": "24 hours"},
      {"condition": "pin_required", "value": true},
      {"condition": "max_value", "value": "1000"}
    ]
  }
}
\```

**Response:**
\```json
{
  "update_request_id": "unique_request_id",
  "status": "pending",
  "new_policy": {
    "policy_id": "unique_policy_id",
    "rules": [
      {"condition": "wait_time", "value": "24 hours"},
      {"condition": "pin_required", "value": true},
      {"condition": "max_value", "value": "1000"}
    ]
  },
  "requested_at": "timestamp"
}
\```

#### [POST] /vaults/{vault_id}/transaction_policies/{update_request_id}/approve

Submit signatures to approve a transaction policy update request. The update will only take effect if all parties of the vault have signed the request.

**Request:**
\```json
{
  "signatures": [
    {"party_id": "party_1_id", "signature": "signature_data"},
    {"party_id": "party_2_id", "signature": "signature_data"}
  ]
}
\```

**Response:**
\```json
{
  "status": "approved",
  "new_policy": {
    "policy_id": "unique_policy_id",
    "rules": [
      {"condition": "wait_time", "value": "24 hours"},
      {"condition": "pin_required", "value": true},
      {"condition": "max_value", "value": "1000"}
    ]
  },
  "approved_at": "timestamp"
}
\```

#### [GET] /vaults/{vault_id}/transaction_policies/history

Retrieve the history of transaction policies for a specific vault. This allows the wallet to notify if the policy has changed. Wallets can store the policy locally and compare it with the latest policy to determine if it has changed.

**Response:**
\```json
{
  "policies": [
    {
      "policy_id": "policy_id_1",
      "rules": [
        {"condition": "wait_time", "value": "24 hours"},
        {"condition": "pin_required", "value": true},
        {"condition": "max_value", "value": "1000"}
      ],
      "status": "inactive",
      "created_at": "timestamp",
      "modified_at": "timestamp"
    },
    {
      "policy_id": "policy_id_2",
      "rules": [
        {"condition": "wait_time", "value": "12 hours"},
        {"condition": "pin_required", "value": false},
        {"condition": "max_value", "value": "500"}
      ],
      "status": "inactive",
      "created_at": "timestamp",
      "modified_at": "timestamp"
    }
  ]
}
\```

### Transaction Signing

#### [POST] /vaults/{vault_id}/transactions

Request the signing of a transaction. This spins up a background job (using `cmd/worker` with `asynq`) to join the TSS-coordinator if the transaction policy allows it; otherwise, it will throw an error.

**Request:**
\```json
{
  "transaction_id": "unique_transaction_id",
  "transaction_details": {
    "amount": 100,
    "recipient": "recipient_address"
  },
  "policy_id": "unique_policy_id",
  "session_id": "unique_session_id",
  "pin_code": "user_provided_pin_code"
}
\```

**Response (initial):**
\```json
{
  "transaction_id": "unique_transaction_id",
  "status": "pending",
  "required_signatures": [
    {"party_id": "party_1_id", "status": "waiting"},
    {"party_id": "party_2_id", "status": "waiting"},
    {"party_id": "Vultisigner", "status": "waiting"}
  ],
  "created_at": "timestamp"
}
\```

#### [GET] /vaults/{vault_id}/transactions/{transaction_id}

View the status of a transaction signing job.

**Response:**
\```json
{
  "transaction_id": "unique_transaction_id",
  "status": "in_progress", // or "completed", "failed" (TBD)
  "required_signatures": [
    {"party_id": "party_1_id", "status": "signed"},
    {"party_id": "party_2_id", "status": "waiting"},
    {"party_id": "Vultisigner", "status": "waiting"}
  ],
  "updated_at": "timestamp"
}
\```

**Response (after keysign completion):**
\```json
{
  "transaction_id": "unique_transaction_id",
  "status": "completed",
  "signed_message": {
    "body": "signed_message_body",
    "hash": "signed_message_hash"
  },
  "required_signatures": [
    {"party_id": "party_1_id", "status": "signed"},
    {"party_id": "party_2_id", "status": "signed"},
    {"party_id": "Vultisigner", "status": "signed"}
  ],
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
\```

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

### Monitoring Job Status

Both key generation and transaction signing jobs can be monitored via the `GET` endpoints in `cmd/vultisigner`.

### Binance-TSS and TSS Coordinator

Binance-TSS involves multiple parties participating in cryptographic operations. The TSS Coordinator facilitates communication between parties, especially useful for mobile devices that typically do not allow inbound connections. The session ID ensures isolation of different vault sessions.

### Encryption

The encryption package handles the encryption and decryption of sensitive data fields, ensuring secure storage.

#### Methods

- **Encrypt**
  - Encrypts plaintext data using AES encryption with GCM mode.
  - Generates a nonce for each encryption operation.
  - Uses a password from the server configuration for encryption.

- **Decrypt**
  - Decrypts ciphertext data encrypted by the Encrypt method.
  - Requires the same password used for encryption.

#### Usage

1. **Encryption of Sensitive Fields**
   - Sensitive fields (e.g., `Xi` fields in `EcdsaLocalData` and `EddsaLocalData`) are encrypted before storage.
   - Ensures data security even if the database is compromised.

2. **Decryption of Sensitive Fields**
   - Converts encrypted data back to plaintext when accessed.
   - Ensures secure and controlled access to sensitive data.

#### Secure Storage Process

- **Vault Share Encryption**
  - Validate sensitive fields.
  - Encrypt `Xi` fields using the `Encrypt` method.
  - Store the encrypted data in the database.
  - Ensures sensitive information remains protected even if the database is compromised.
