// Package art is the painted tree the picture embeds.
// A PNG is a sheet only when a JSON manifest with the same path sits beside it.
// The style board under style/ has no manifests, so the game does not draw it.
package art

import "embed"

// Files is every file under assets/art. A later sheet dropped next to a
// manifest is picked up on the next build; this pattern is the whole tree.
//
//go:embed *
var Files embed.FS
