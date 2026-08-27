// Package migrations embute os arquivos SQL de migração do banco.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
