up:
	@docker compose up -d --remove-orphans;

down:
	@docker compose down

verifier-server:
	VS_CONFIG_NAME=config-verifier go run cmd/vultisigner/main.go

plugin-server:
	VS_CONFIG_NAME=config-plugin go run cmd/vultisigner/main.go

verifier-worker:
	VS_CONFIG_NAME=config-verifier go run cmd/worker/main.go

plugin-worker:
	VS_CONFIG_NAME=config-plugin go run cmd/worker/main.go
