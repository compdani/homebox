package migrations

import (
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(upHomeboxCollections, downHomeboxCollections)
}

func groupRules() (list, view, create, update, delete string) {
	return collections.GroupScopeRule, collections.GroupScopeRule, collections.GroupMemberWriteRule, collections.GroupMemberWriteRule, collections.GroupOwnerWriteRule
}

func upHomeboxCollections(app core.App) error {
	groups := core.NewBaseCollection(collections.Groups)
	groups.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "currency", Required: true, Max: 10},
	)
	addTimestamps(&groups.Fields)
	groups.ListRule = types.Pointer(collections.GroupMemberViewRule)
	groups.ViewRule = types.Pointer(collections.GroupMemberViewRule)
	groups.UpdateRule = types.Pointer(collections.GroupOwnerUpdateRule)
	if err := app.Save(groups); err != nil {
		return err
	}

	users, err := app.FindCollectionByNameOrId(collections.Users)
	if err != nil {
		return err
	}
	users.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.SelectField{
			Name:      "role",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"user", "owner"},
		},
		&core.RelationField{
			Name:          "group",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  groups.Id,
			CascadeDelete: true,
		},
	)
	users.ListRule = types.Pointer(`@request.auth.id != "" && id = @request.auth.id`)
	users.ViewRule = types.Pointer(`@request.auth.id != "" && id = @request.auth.id`)
	users.CreateRule = types.Pointer("")
	users.UpdateRule = types.Pointer(collections.UserSelfUpdateRule)
	users.DeleteRule = types.Pointer(`@request.auth.id != "" && id = @request.auth.id`)
	if err := app.Save(users); err != nil {
		return err
	}

	locations := core.NewBaseCollection(collections.Locations)
	list, view, create, update, del := groupRules()
	locations.ListRule = types.Pointer(list)
	locations.ViewRule = types.Pointer(view)
	locations.CreateRule = types.Pointer(create)
	locations.UpdateRule = types.Pointer(update)
	locations.DeleteRule = types.Pointer(del)
	locations.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 1000},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&locations.Fields)
	if err := app.Save(locations); err != nil {
		return err
	}
	locations.Fields.Add(&core.RelationField{
		Name:         "parent",
		MaxSelect:    1,
		CollectionId: locations.Id,
	})
	if err := app.Save(locations); err != nil {
		return err
	}

	labels := core.NewBaseCollection(collections.Labels)
	labels.ListRule = types.Pointer(list)
	labels.ViewRule = types.Pointer(view)
	labels.CreateRule = types.Pointer(create)
	labels.UpdateRule = types.Pointer(update)
	labels.DeleteRule = types.Pointer(del)
	labels.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 1000},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&labels.Fields)
	if err := app.Save(labels); err != nil {
		return err
	}

	items := core.NewBaseCollection(collections.Items)
	items.ListRule = types.Pointer(list)
	items.ViewRule = types.Pointer(view)
	items.CreateRule = types.Pointer(create)
	items.UpdateRule = types.Pointer(update)
	items.DeleteRule = types.Pointer(del)
	items.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 1000},
		&core.TextField{Name: "import_ref", Max: 100},
		&core.TextField{Name: "notes", Max: 1000},
		&core.NumberField{Name: "quantity", OnlyInt: true},
		&core.BoolField{Name: "insured"},
		&core.BoolField{Name: "archived"},
		&core.NumberField{Name: "asset_id", OnlyInt: true},
		&core.TextField{Name: "serial_number", Max: 255},
		&core.TextField{Name: "model_number", Max: 255},
		&core.TextField{Name: "manufacturer", Max: 255},
		&core.BoolField{Name: "lifetime_warranty"},
		&core.DateField{Name: "warranty_expires"},
		&core.TextField{Name: "warranty_details", Max: 1000},
		&core.DateField{Name: "purchase_time"},
		&core.TextField{Name: "purchase_from"},
		&core.NumberField{Name: "purchase_price"},
		&core.DateField{Name: "sold_time"},
		&core.TextField{Name: "sold_to"},
		&core.NumberField{Name: "sold_price"},
		&core.TextField{Name: "sold_notes", Max: 1000},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
		&core.RelationField{
			Name:         "location",
			MaxSelect:    1,
			CollectionId: locations.Id,
		},
		&core.RelationField{
			Name:         "labels",
			MaxSelect:    999,
			CollectionId: labels.Id,
		},
	)
	addTimestamps(&items.Fields)
	if err := app.Save(items); err != nil {
		return err
	}
	items.Fields.Add(&core.RelationField{
		Name:         "parent",
		MaxSelect:    1,
		CollectionId: items.Id,
	})
	if err := app.Save(items); err != nil {
		return err
	}

	itemFields := core.NewBaseCollection(collections.ItemFields)
	itemFields.ListRule = types.Pointer(list)
	itemFields.ViewRule = types.Pointer(view)
	itemFields.CreateRule = types.Pointer(create)
	itemFields.UpdateRule = types.Pointer(update)
	itemFields.DeleteRule = types.Pointer(del)
	itemFields.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 1000},
		&core.SelectField{
			Name:      "type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"text", "number", "boolean", "time"},
		},
		&core.TextField{Name: "text_value", Max: 500},
		&core.NumberField{Name: "number_value", OnlyInt: true},
		&core.BoolField{Name: "boolean_value"},
		&core.DateField{Name: "time_value"},
		&core.RelationField{
			Name:          "item",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  items.Id,
			CascadeDelete: true,
		},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&itemFields.Fields)
	if err := app.Save(itemFields); err != nil {
		return err
	}

	attachments := core.NewBaseCollection(collections.Attachments)
	attachments.ListRule = types.Pointer(list)
	attachments.ViewRule = types.Pointer(view)
	attachments.CreateRule = types.Pointer(create)
	attachments.UpdateRule = types.Pointer(update)
	attachments.DeleteRule = types.Pointer(del)
	attachments.Fields.Add(
		&core.SelectField{
			Name:      "type",
			MaxSelect: 1,
			Values:    []string{"photo", "manual", "warranty", "attachment", "receipt"},
		},
		&core.BoolField{Name: "primary"},
		&core.TextField{Name: "title", Max: 255},
		&core.FileField{Name: "file", Required: true, MaxSelect: 1, MaxSize: 52428800},
		&core.RelationField{
			Name:          "item",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  items.Id,
			CascadeDelete: true,
		},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&attachments.Fields)
	if err := app.Save(attachments); err != nil {
		return err
	}

	maintenance := core.NewBaseCollection(collections.Maintenance)
	maintenance.ListRule = types.Pointer(list)
	maintenance.ViewRule = types.Pointer(view)
	maintenance.CreateRule = types.Pointer(create)
	maintenance.UpdateRule = types.Pointer(update)
	maintenance.DeleteRule = types.Pointer(del)
	maintenance.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.TextField{Name: "description", Max: 2500},
		&core.DateField{Name: "date"},
		&core.DateField{Name: "scheduled_date"},
		&core.NumberField{Name: "cost"},
		&core.RelationField{
			Name:          "item",
			Required:      true,
			MaxSelect:     1,
			CollectionId:  items.Id,
			CascadeDelete: true,
		},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&maintenance.Fields)
	if err := app.Save(maintenance); err != nil {
		return err
	}

	notifiers := core.NewBaseCollection(collections.Notifiers)
	notifiers.ListRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	notifiers.ViewRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	notifiers.CreateRule = types.Pointer(`@request.auth.id != ""`)
	notifiers.UpdateRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	notifiers.DeleteRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	notifiers.Fields.Add(
		&core.TextField{Name: "name", Required: true, Max: 255},
		&core.URLField{Name: "url", Required: true},
		&core.BoolField{Name: "is_active"},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
		&core.RelationField{
			Name:         "user",
			Required:     true,
			MaxSelect:    1,
			CollectionId: users.Id,
		},
	)
	addTimestamps(&notifiers.Fields)
	if err := app.Save(notifiers); err != nil {
		return err
	}

	invitations := core.NewBaseCollection(collections.GroupInvitations)
	invitations.CreateRule = types.Pointer(collections.InvitationOwnerWriteRule)
	invitations.DeleteRule = types.Pointer(collections.InvitationOwnerWriteRule)
	invitations.Fields.Add(
		&core.TextField{Name: "token_hash", Required: true, Max: 128},
		&core.DateField{Name: "expires_at", Required: true},
		&core.NumberField{Name: "uses", OnlyInt: true},
		&core.RelationField{
			Name:         "group",
			Required:     true,
			MaxSelect:    1,
			CollectionId: groups.Id,
		},
	)
	addTimestamps(&invitations.Fields)
	return app.Save(invitations)
}

func downHomeboxCollections(app core.App) error {
	names := []string{
		collections.GroupInvitations,
		collections.Notifiers,
		collections.Maintenance,
		collections.Attachments,
		collections.ItemFields,
		collections.Items,
		collections.Labels,
		collections.Locations,
		collections.Groups,
	}
	for _, name := range names {
		c, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			continue
		}
		if err := app.Delete(c); err != nil {
			return err
		}
	}
	return nil
}
