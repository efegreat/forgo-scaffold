// Package fixtures holds the SQL fixtures of the tests, by scenario, embedded so that every test package
// reads them wherever it runs. Load them with support.Fixture.
package fixtures

import "embed"

//go:embed */*.sql
var FS embed.FS
