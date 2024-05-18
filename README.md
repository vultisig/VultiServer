# Vultisigner

## [POST] /keygen

A user can call this to request the operator of the Vultisigner instance to join the keygen session with them, so that the operator will own one of the share's in the vaults.

Request:

```json
{
  "parties": "iphone,pc,vultisigner",
  "session": "782378237823478242",
  "chain_code": "1"
}
```

Response:

```json
{
  "parties": "iphone,pc,vultisigner",
  "session": "782378237823478242",
  "chain_code": "1",
  "public_key": "0x123...."
}
```
