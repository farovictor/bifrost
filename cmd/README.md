# Command Line Interface

The `bifrost` CLI in this directory interacts with the running HTTP API.
Common commands include:

- `issue` – create a virtual key
- `revoke` – revoke a virtual key
- `service-add` and `service-delete` – manage upstream services
- `rootkey-add`, `rootkey-update`, `rootkey-delete` – manage root keys
- `user-add` – create an API user
- `migrate` – apply SQL migrations in `migrations/`
- `init-admin` – create an admin user and organization in the database
  (uses `BIFROST_ADMIN_ID`, `BIFROST_ADMIN_API_KEY`,
  `BIFROST_ADMIN_ORG_ID`, and `BIFROST_ADMIN_ORG_NAME` if set)

All commands talk to `http://localhost:3333` by default. Use `--addr` to
override the API address.
