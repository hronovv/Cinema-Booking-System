run:
	@docker compose up -d && go run ./cmd

down:
	@docker compose down

test:
	@go test ./... -v -count=1
