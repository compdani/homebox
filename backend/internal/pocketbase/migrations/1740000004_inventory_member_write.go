package migrations

import (
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(upInventoryMemberWrite, downInventoryMemberWrite)
}

func upInventoryMemberWrite(app core.App) error {
	names := []string{
		collections.Locations,
		collections.Labels,
		collections.Items,
		collections.Products,
		collections.ItemFields,
		collections.Attachments,
		collections.Maintenance,
	}
	for _, name := range names {
		col, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			return err
		}
		col.CreateRule = types.Pointer(collections.GroupMemberWriteRule)
		col.UpdateRule = types.Pointer(collections.GroupMemberWriteRule)
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}

func downInventoryMemberWrite(app core.App) error {
	names := []string{
		collections.Locations,
		collections.Labels,
		collections.Items,
		collections.Products,
		collections.ItemFields,
		collections.Attachments,
		collections.Maintenance,
	}
	for _, name := range names {
		col, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			continue
		}
		col.CreateRule = types.Pointer(collections.GroupOwnerWriteRule)
		col.UpdateRule = types.Pointer(collections.GroupOwnerWriteRule)
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}
