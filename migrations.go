// Package example holds what every deployable reads from the root of the module: the SQL migrations,
// embedded so that the binaries and the tests need no files on disk.
package example

import "embed"

// Migrations holds the migration files of each database, under migrations/postgres and migrations/mysql.
//
//go:embed migrations
var Migrations embed.FS
