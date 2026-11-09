// Package migrations nhúng các file .sql vào binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
