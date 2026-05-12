// Package migrations embeds the SQL migration files so they can be applied at
// startup without depending on the on-disk layout.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
