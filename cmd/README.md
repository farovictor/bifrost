# Command Line Interface

The `bifrost` CLI in this directory interacts with the running HTTP API.
Common commands include:

- `issue` – create a virtual key
- `revoke` – revoke a virtual key
- `service-add` and `service-delete` – manage upstream services
- `rootkey-add`, `rootkey-update`, `rootkey-delete` – manage root keys
- `user-add` – create an API user

All commands talk to `http://localhost:3333` by default. Use `--addr` to
override the API address.
