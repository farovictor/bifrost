package pq

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

// pqDriver is a stub implementing driver.Driver.
type pqDriver struct{}

func (pqDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("stub postgres driver")
}

func init() {
	sql.Register("postgres", pqDriver{})
}
