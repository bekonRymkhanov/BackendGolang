# If the first argument is "migrate"...
ifeq (migrate, $(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate"
  MIGRATE_ARGS := $(wordlist 2, $(words $(MAKECMDGOALS)), $(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_ARGS):;@:)
endif

migrate:
	go run cmd/migrate/main.go \
	-source file://migration \
	-database postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable \
	$(MIGRATE_ARGS)

docker-compose-up:
	docker-compose -f docker-compose.auth.yml up
