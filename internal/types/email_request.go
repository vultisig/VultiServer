package types

type EmailRequest struct {
	Email        string `json:"email"`
	FileName     string `json:"file_name"`
	BackupKey    string `json:"backup_key"`
	BackupDigest string `json:"backup_digest"`
	// FileContent is only set by tasks enqueued before BackupKey existed.
	FileContent string `json:"file_content,omitempty"`
	VaultName   string `json:"vault_name"`
	Code        string `json:"code"`
}

type ResendVaultShareEmailRequest struct {
	Email        string `json:"email"`
	FileName     string `json:"file_name"`
	BackupKey    string `json:"backup_key"`
	BackupDigest string `json:"backup_digest"`
	// FileContent is only set by tasks enqueued before BackupKey existed.
	FileContent string `json:"file_content,omitempty"`
	VaultName   string `json:"vault_name"`
}
