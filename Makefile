gen-vault-dirs:
	mkdir -p tmp/vultisigner/server/vaults
	mkdir -p tmp/vultisigner/plugin/vaults
	sudo chmod 777 tmp/vultisigner/server/vaults
	sudo chmod 777 tmp/vultisigner/plugin/vaults

up:
	@docker compose up -d --remove-orphans;

down:
	@docker compose down

# start servers/workers in this order

signer-server:
	VS_CONFIG_NAME=config-server go run cmd/vultisigner/main.go

plugin-server:
	VS_CONFIG_NAME=config-plugin go run cmd/vultisigner/main.go

signer-worker:
	VS_CONFIG_NAME=config-server go run cmd/worker/main.go

plugin-worker:
	VS_CONFIG_NAME=config-plugin go run cmd/worker/main.go
