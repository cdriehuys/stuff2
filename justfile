default:
    just --list

# Generate application code
[group('app')]
generate:
    go generate ./...

# Run all tests
[group('app')]
test: generate
    go test -v -count=1 ./...

# Open a database shell
[group('database')]
db-shell:
    psql --host "${POSTGRES_HOSTNAME}" --username "${POSTGRES_USER}"

export TERN_MIGRATIONS := './migrations'

# Migrate the database to the latest version
[group('database')]
migrate: (_tern "migrate")

# Migration targets may be a migration number, a positive or negative delta, or
# 0 to revert all migrations.
#
# Migrate to a particular state
[group('database')]
migrate-to target: (_tern "migrate" "--destination" target)

# Create a new migration
[group('database')]
new-migration name: (_tern "new" name)

# Run a generic `tern` command
[group('database')]
_tern +ARGS:
    go tool tern {{ARGS}}
