# Command Line Interface

The `bifrost` CLI in this directory interacts with the running HTTP API. Common
commands are listed below:

| Command | Description |
|---------|-------------|
| `issue` | create a virtual key |
| `revoke` | revoke a virtual key |
| `service-add` | add an upstream service |
| `service-delete` | delete an upstream service |
| `rootkey-add` | add a root key |
| `rootkey-update` | update a root key |
| `rootkey-delete` | delete a root key |
| `user-add` | create an API user |
| `migrate` | apply SQL migrations in `migrations/` |
| `init-admin` | create an admin user in the database (uses `BIFROST_ADMIN_API_KEY`, `BIFROST_ADMIN_NAME`, and `BIFROST_ADMIN_EMAIL` if set) |

All commands talk to `http://localhost:3333` by default. Use `--addr` to
override the API address.
