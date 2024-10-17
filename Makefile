# Include environment variables
include .env
export

# Define Phony targets
.PHONY: migrate.up migrate.up.all migrate.down migrate.down.all migration migrate.force test

# Target to apply migration
migrate.up:
		migrate -path=$(MIGRATIONS_ROOT) -database $(DATABASE_URL) up $(n)

# Target to apply all migrations
migrate.up.all:
		migrate -path=$(MIGRATIONS_ROOT) -database $(DATABASE_URL) up

# Target to rollback a migration
migrate.down:
		migrate -path=$(MIGRATIONS_ROOT) -database $(DATABASE_URL) down $(n)

# Target to rollback all migrations
migrate.down.all:
		migrate -path=$(MIGRATIONS_ROOT) -database $(DATABASE_URL) down -all

# Target to create a new migration
migration:
		migrate create -seq -ext=.sql -dir=$(MIGRATIONS_ROOT) $(n)

# Target to force a migration version
migrate.force:
		migrate -path=$(MIGRATIONS_ROOT) -database=$(DATABASE_URL) force $(n)

# Test target
test:
		@echo "Test target is working"
