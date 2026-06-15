package migrations

import (
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/pocketbase/core"

	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(upAddAutodateFields, downAddAutodateFields)
}

func upAddAutodateFields(app core.App) error {
	return ensureTimestamps(app,
		collections.Groups,
		collections.Locations,
		collections.Labels,
		collections.Items,
		collections.ItemFields,
		collections.Attachments,
		collections.Maintenance,
		collections.Notifiers,
		collections.GroupInvitations,
	)
}

func downAddAutodateFields(app core.App) error {
	return nil
}
