package migrations

import "embed"

// Files embeds versioned upward migrations; running binaries do not rely on cwd.
//
//go:embed *.up.sql
var Files embed.FS
