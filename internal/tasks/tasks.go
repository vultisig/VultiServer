package tasks

const QUEUE_NAME = "vultisigner"
const EMAIL_QUEUE_NAME = "vultisigner:email"
const (
	TypeKeyGeneration     = "key:generation"
	TypeKeySign           = "key:sign"
	TypeEmailVaultBackup  = "key:email:v2"
	TypeReshare           = "key:reshare"
	TypeKeyGenerationDKLS = "key:generationDKLS"
	TypeKeySignDKLS       = "key:signDKLS"
	TypeReshareDKLS       = "key:reshareDKLS"
	TypeMigrate           = "key:migrate"
	TypeImport            = "key:import"
	TypeCreateMldsa              = "key:createMldsa"
	TypeKeygenBatch              = "key:keygenBatch"
	TypeReshareBatch             = "key:reshareBatch"
	TypeImportBatch              = "key:importBatch"
	TypeResendVaultShareEmail    = "key:resendVaultShareEmail:v2"
)

// Pre-v2 email task types carry the backup itself; registered only to drain tasks from older binaries.
const (
	TypeEmailVaultBackupLegacy      = "key:email"
	TypeResendVaultShareEmailLegacy = "key:resendVaultShareEmail"
)
