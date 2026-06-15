package hooks

import (
	"github.com/pocketbase/pocketbase/core"
)

// Register attaches Homebox PocketBase record hooks.
func Register(app core.App) {
	// Group scoping is enforced via collection API rules and client payloads.
	_ = app
}
