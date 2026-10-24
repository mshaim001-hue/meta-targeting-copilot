.PHONY: dev api web test pipeline tidy

dev:
	docker compose up db -d
	@echo "API:  make api"
	@echo "WEB:  make web"

api:
	cd backend && DATABASE_URL=$${DATABASE_URL:-postgres://copilot:copilot@localhost:5432/copilot?sslmode=disable} go run ./cmd/api

web:
	cd frontend && npm run dev

pipeline:
	cd backend && go run ./cmd/pipeline --product "Fitness club" --city Astana --country Kazakhstan --audience "People who already exercise regularly."

test:
	cd backend && go test ./...

tidy:
	cd backend && go mod tidy
