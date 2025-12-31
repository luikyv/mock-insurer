.PHONY: setup-dev setup-cs cs-tests run run-with-cs keys generate migration build lint test test-coverage cs-tests

ORG_ID="00000000-0000-0000-0000-000000000000"
SOFTWARE_ID="11111111-1111-1111-1111-111111111111"
CS_VERSION="2d752b24d94fcbddd19451293f5203b10409de15"

setup-dev:
	@go mod download
	@pre-commit install

# Clone and build the Open Insurance Conformance Suite.
setup-cs:
	@if [ ! -d "conformance/suite" ]; then \
	  echo "Cloning open insurance conformance suite repository..."; \
	  cd conformance; \
	  git clone https://gitlab.com/raidiam-conformance/open-insurance/open-insurance-brasil.git suite; \
	fi
	
	@if [ ! -d "conformance/venv" ]; then \
	  python3 -m venv conformance/venv; \
	  . ./conformance/venv/bin/activate; \
	  python3 -m pip install httpx pyparsing; \
	fi

	@cd conformance/suite && git checkout $(CS_VERSION)
	@docker compose run cs-builder

run:
	@docker compose up

# Start Mock Insurer along with the Open Finance Conformance Suite.
run-with-cs:
	@docker compose --profile conformance up

# Generate certificates, private keys, and JWKS files for both the server and clients.
keys:
	@go run cmd/keymaker/main.go --org_id=$(ORG_ID) --software_id=$(SOFTWARE_ID) --keys_dir=./keys

generate:
	@go generate ./...

migration:
	@docker compose run migration

build:
	@docker compose build

lint:
	@golangci-lint run ./...

test-db:
	@docker compose up psql-test

test:
	@go test ./internal/...

test-coverage:
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html="coverage.out" -o coverage.html
	@echo "Total Coverage: `go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'` %"

cs-tests:
	@conformance/venv/bin/python conformance/run-test-plan.py \
		"Insurance auto api test v1.4.0" ./conformance/config.json \
		"Insurance customer personal api test-v1.6.0" ./conformance/config.json \
		"capitalization-title_test-plan_v1.5.0" ./conformance/config.json \
		"financial-assistance_test-plan_v1n3" ./conformance/config.json \
		"Insurance acceptance and branches abroad api test V1.4.0" ./conformance/config.json \
		"Insurance financial risks api test-V1n4" ./conformance/config.json \
		"Insurance Housing API test V1.4.0" ./conformance/config.json \
		"life-pension_test-plan_v1n5" ./conformance/config.json \
		"Insurance patrimonial api test-v1.5.0" ./conformance/config.json \
		"quote-auto_test-plan-v1.10" ./conformance/config.json \
		--expected-skips-file ./conformance/expected_skips.json \
		--expected-failures-file ./conformance/expected_failures.json \
		--export-dir ./conformance/results \
		--verbose
