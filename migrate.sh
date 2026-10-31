#!/bin/bash

set -e

# Default to "up" if no argument is provided
COMMAND="${1:-up}"

DB_URL="postgresql://postgres:secret@localhost:5431/taskify_v3"
MIGRATION_DIR="db/migration"

echo "Running goose migrations ($COMMAND)..."
goose -dir "$MIGRATION_DIR" postgres "$DB_URL" "$COMMAND"