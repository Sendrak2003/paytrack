//go:build postgres

package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// openDialector (postgres build) supports both SQLite and Postgres.
// Enable with: go build -tags postgres ./cmd/server
// Requires: go get gorm.io/driver/postgres
func openDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite", "":
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unknown DB_DRIVER %q", driver)
	}
}
