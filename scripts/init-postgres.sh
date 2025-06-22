#!/bin/bash
set -e

# Apply migrations if present
for f in /docker-entrypoint-initdb.d/migrations/*.sql; do
    [ -f "$f" ] || continue
    echo "Applying $f"
    psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$f"
done
