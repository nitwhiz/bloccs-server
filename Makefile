.PHONY: dev
dev:
	docker run \
		--rm \
		-it \
		--network="host" \
		--volume "$(shell pwd)/:/app" \
		--workdir "/app" \
		golang:1.18.2 \
		go run "./cmd/server"

.PHONY: dev_race
dev_race:
	docker run \
		--rm \
		-it \
		--network="host" \
		--volume "$(shell pwd)/:/app" \
		--workdir "/app" \
		golang:1.18.2 \
		go run -race "./cmd/server"

.PHONY: test_e2e
test_e2e:
	docker compose \
		--file docker-compose.e2e.yml \
		up --exit-code-from test_runner

	docker compose \
		--file docker-compose.e2e.yml \
		down --remove-orphans

.PHONY: test_unit
test_unit:
	docker compose \
		--file docker-compose.unit.yml \
		up --exit-code-from test_runner

	docker compose \
		--file docker-compose.unit.yml \
		down --remove-orphans

.PHONY: test
test: test_unit test_e2e
