# Database Migrations

This directory contains plain SQL migration files. Apply them sequentially to initialize the database schema.

Run each file against your database in order, using the tool of your choice. For example with `psql`:

```bash
psql $DATABASE_URL -f migrations/001_create_organizations.sql
psql $DATABASE_URL -f migrations/002_create_org_memberships.sql
```

Any migration tool that executes SQL scripts in order (like `migrate` or `goose`) will also work.

After the schema is in place run `bifrost init-admin` to create the initial admin
user. Set `BIFROST_ADMIN_API_KEY`, `BIFROST_ADMIN_NAME`, and `BIFROST_ADMIN_EMAIL`
to override the defaults.
The command prints the resulting API key so you can store it securely.

