package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

func addTimestamps(fields *core.FieldsList) {
	fields.Add(
		&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		},
		&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		},
	)
}

func collectionHasField(col *core.Collection, name string) bool {
	for _, field := range col.Fields {
		if field.GetName() == name {
			return true
		}
	}
	return false
}

func ensureTimestamps(app core.App, names ...string) error {
	for _, name := range names {
		col, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			return err
		}
		if collectionHasField(col, "created") {
			continue
		}
		addTimestamps(&col.Fields)
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}
