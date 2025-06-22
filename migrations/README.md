# Database Migrations

This directory contains plain SQL migration files. Apply them sequentially to initialize the database schema.

Run each file against your database in order, using the tool of your choice. For example with `psql`:

```bash
psql $DATABASE_URL -f migrations/001_create_organizations.sql
psql $DATABASE_URL -f migrations/002_create_org_memberships.sql
psql $DATABASE_URL -f migrations/003_create_users.sql
psql $DATABASE_URL -f migrations/004_create_root_keys.sql
psql $DATABASE_URL -f migrations/005_create_services.sql
psql $DATABASE_URL -f migrations/006_create_virtual_keys.sql
```

Any migration tool that executes SQL scripts in order (like `migrate` or `goose`) will also work.

