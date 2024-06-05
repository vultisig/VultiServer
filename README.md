# Vultisigner

## cmd/vultisigner

### Vault Management

#### [POST] /vault

Create a new vault with specified parties. This does not join a key generation yet. Only creates it.

A vault can never be updated to have new signers, a vault is always considered valid, if participants change, we consider that a new vault, as the old vault can technically still be used to sign.

**Request:**
```js
{
  "name": "My Vault",
  "hex_chain_code": "80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ]
}
```

**Response (initial):**
```js
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "hexChainCode": "80871c0f885f953e5206e461630a9222148797e66276a83224c7b9b0f75b3ec0",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "created_at": "timestamp"
}
```

#### [POST] /vault/{vault_id}/keygen
After the vault is created, you can call this endpoint, this will start the key generation process, spinning up a background job (using `cmd/worker` with `asynq`) to join relay keygen session.

This endpoint is meant to return the session_id, which allows the other parties to join the keygen session.

**Request:**
```js
// nothing
```

**Response:**
```js
{
  "vault_id": "unique_vault_id",
  "session_id": "unique_session_id",
}
```

#### [GET] /vault/{vault_id}/keygen
You can retrieve the status of a keygen session here

**Response:**
```js
{
  "vault_id": "unique_vault_id",
  "session_id": "unique_session_id",
  "status": "pending", // or "in_progress", "completed", "failed", "prime-generating"
}
```

#### [GET] /vault/{vault_id}

**Response (after creation):**
```js
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "policy": {},
  "shares": {},
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Response (after keygen completion):**
```js
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "policy": {},
  "shares": {
    "ecdsa": [
      {"party_id": "party_1_id"},
      {"party_id": "party_2_id"},
      {"party_id": "Vultisigner"
    ],
    "eddsa": [
      {"party_id": "party_1_id"},
      {"party_id": "party_2_id"},
      {"party_id": "Vultisigner"}
    ]
  },
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

**Response (after keygen completion and has active transaction policy configured):**
```js
{
  "vault_id": "unique_vault_id",
  "name": "My Vault",
  "parties": [
    "party_1_id",
    "party_2_id",
    "Vultisigner"
  ],
  "policy": {
    "policy_id": "unique_policy_id",
    "rules": [
      {"condition": "wait_time", "value": "24 hours"},
      {"condition": "pin_required", "value": true},
      {"condition": "max_value", "value": "1000"}
    ]
  },
  "shares": {
    "ecdsa": [
      {"party_id": "party_1_id" },
      {"party_id": "party_2_id" },
      {"party_id": "Vultisigner" }
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
```

### Transaction Policy Management

#### [POST] /vault/{vault_id}/transaction_policies/request_update

Request an update to the transaction policy for a specific vault. This generates data that other parties must sign to approve the update. They should sign the contents of `new_policy` and send the signatures to the `/vault/{vault_id}/transaction_policies/{update_request_id}/approve` endpoint for it to take effect.

**Request:**
```js
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
```

**Response:**
```js
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
```

#### [GET] /vault/{vault_id}/transaction_policies/{update_request_id}

Retrieve the details of a transaction policy update request. This allows the wallet to display the details of the request to the user and prompt them to sign it.

**Response:**
```js
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
```

#### [POST] /vault/{vault_id}/transaction_policies/{update_request_id}/approve

Submit signatures to approve a transaction policy update request. The update will only take effect if a vault minimum threshold of signatures is reached. We will never be one of the signers.

**Request:**
```js
{
  "signatures": [
    {"party_id": "party_1_id", "signature": "signature_data"},
    {"party_id": "party_2_id", "signature": "signature_data"}
  ]
}
```

**Response:**
```js
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
```

#### [GET] /vault/{vault_id}/transaction_policies/history

Retrieve the history of transaction policies for a specific vault. This allows the wallet to notify if the policy has changed. Wallets can store the policy locally and compare it with the latest policy to determine if it has changed.

**Response:**
```js
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
```

#### [GET] /vault/{vault_id}/transaction_policies/{policy_id}

Retrieve a specific transaction policy for a specific vault.

**Response:**
```js
{
      "policy_id": "policy_id_1",
      "rules": [
        {"condition": "wait_time", "value": "24 hours"},
        {"condition": "pin_required", "value": true},
        {"condition": "max_value", "value": "1000"}
      ],
      "status": "inactive", // or "active"
      "created_at": "timestamp",
      "modified_at": "timestamp"
}
```

### Transaction Signing

#### [POST] /vault/{vault_id}/transactions

Request the signing of a transaction. This spins up a background job (using `cmd/worker` with `asynq`) to join the TSS-coordinator if the transaction policy allows it; otherwise, it will throw an error.

**Request:**
```js
{
  "transaction_details": {
    "body": "transaction_body",
  },
  "policy_id": "unique_policy_id",
  "session_id": "unique_session_id",
  "pin_code": "user_provided_pin_code" // optional
}
```

**Response (initial):**
```js
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
```

**Response (not allowed):**
```js
{
  "transaction_id": "unique_transaction_id",
  "status": "unauthorized",
  "unauthorized_reason": "timelock_active", // or "pin_required", "max_value_exceeded"
  "required_signatures": [
    {"party_id": "party_1_id", "status": "waiting"},
    {"party_id": "party_2_id", "status": "waiting"},
    {"party_id": "Vultisigner", "status": "waiting"}
  ],
  "created_at": "timestamp"
}
```

#### [GET] /vault/{vault_id}/transactions/{transaction_id}

View the status of a transaction signing job.

**Response:**
```js
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
```

**Response (after keysign completion):**
```js
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
```

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
