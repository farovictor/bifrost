# Database Connection

The `database` package exposes helpers to create a `*gorm.DB` instance.
It opens the connection using the provided DSN and verifies that the
underlying database is reachable.
