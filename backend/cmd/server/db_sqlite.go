//go:build !postgres

package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// openDialector (default build) supports SQLite only — zero extra dependencies.
// To enable Postgres, build with: go build -tags postgres ./cmd/server
func openDialector(driver, dsn string) (gorm.Dialector, error) {
	switch driver {
	case "sqlite", "":
		return sqlite.Open(dsn), nil
	case "postgres":
		return nil, fmt.Errorf("postgres support not compiled in; rebuild with: go build -tags postgres")
	default:
		return nil, fmt.Errorf("unknown DB_DRIVER %q", driver)
	}
}
