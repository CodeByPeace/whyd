// Package rules holds the built-in rule files, embedded in the binary.
package rules

import "embed"

// Builtin contains every *.toml file of this directory.
//
//go:embed *.toml
var Builtin embed.FS
