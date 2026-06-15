package migrations

import (
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(upProducts, downProducts)
}

func upProducts(app core.App) error {
	groups, err := app.FindCollectionByNameOrId(collections.Groups)
	if err != nil {
		return err
	}

	items, err := app.FindCollectionByNameOrId(collections.Items)
	if err != nil {
		return err
	}

	products := core.NewBaseCollection(collections.Products)
	list, view, create, update, del := groupRules()
	products.ListRule = types.Pointer(list)
	products.ViewRule = types.Pointer(view)
	products.CreateRule = types.Pointer(create)
	products.UpdateRule = types.Pointer(update)
	products.DeleteRule = types.Pointer(del)
	products.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 1000},
		&core.TextField{Name: "manufacturer", Max: 255},
		&core.TextField{Name: "model_number", Max: 255},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&products.Fields)
	if err := app.Save(products); err != nil {
		return err
	}

	for _, field := range items.Fields {
		if field.GetName() != "name" {
			continue
		}
		if tf, ok := field.(*core.TextField); ok {
			tf.Required = false
		}
	}

	items.Fields.Add(&core.RelationField{
		Name:         "product",
		MaxSelect:    1,
		CollectionId: products.Id,
	})
	return app.Save(items)
}

func downProducts(app core.App) error {
	items, err := app.FindCollectionByNameOrId(collections.Items)
	if err != nil {
		return err
	}
	items.Fields.RemoveByName("product")
	for _, field := range items.Fields {
		if field.GetName() != "name" {
			continue
		}
		if tf, ok := field.(*core.TextField); ok {
			tf.Required = true
		}
	}
	if err := app.Save(items); err != nil {
		return err
	}

	c, err := app.FindCollectionByNameOrId(collections.Products)
	if err != nil {
		return nil
	}
	return app.Delete(c)
}
