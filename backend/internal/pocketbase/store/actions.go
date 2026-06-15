package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func (s *Store) EnsureAssetIDs(ctx context.Context, groupID string) (int, error) {
	items, err := s.findAll(collections.Items, "group = {:gid} && asset_id = 0", dbx.Params{"gid": groupID})
	if err != nil {
		return 0, err
	}

	highest, err := s.highestAssetID(groupID)
	if err != nil {
		return 0, err
	}

	completed := 0
	for _, item := range items {
		highest++
		item.Set("asset_id", highest)
		if err := s.app.Save(item); err != nil {
			return completed, err
		}
		completed++
	}
	return completed, nil
}

func (s *Store) EnsureImportRefs(ctx context.Context, groupID string) (int, error) {
	items, err := s.findAll(collections.Items, "group = {:gid} && import_ref = ''", dbx.Params{"gid": groupID})
	if err != nil {
		return 0, err
	}

	completed := 0
	for _, item := range items {
		ref := uuid.New().String()[:8]
		item.Set("import_ref", ref)
		if err := s.app.Save(item); err != nil {
			return completed, err
		}
		completed++
	}
	return completed, nil
}

func (s *Store) ZeroOutTimeFields(ctx context.Context, groupID string) (int, error) {
	items, err := s.findAll(collections.Items, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return 0, err
	}

	completed := 0
	for _, item := range items {
		changed := false
		for _, field := range []string{"purchase_time", "sold_time", "warranty_expires"} {
			if dt := item.GetDateTime(field); !dt.IsZero() {
				t := dt.Time()
				zeroed := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				item.Set(field, zeroed)
				changed = true
			}
		}
		if changed {
			if err := s.app.Save(item); err != nil {
				return completed, err
			}
			completed++
		}
	}
	return completed, nil
}

func (s *Store) SetPrimaryPhotos(ctx context.Context, groupID string) (int, error) {
	items, err := s.findAll(collections.Items, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return 0, err
	}

	completed := 0
	for _, item := range items {
		attachments, err := s.app.FindRecordsByFilter(
			collections.Attachments,
			"item = {:id} && type = 'photo'",
			"",
			1,
			0,
			dbx.Params{"id": item.Id},
		)
		if err != nil || len(attachments) == 0 {
			continue
		}

		allPhotos, err := s.findAll(collections.Attachments, "item = {:id} && type = 'photo'", dbx.Params{"id": item.Id})
		if err != nil {
			return completed, err
		}

		changed := false
		for _, att := range allPhotos {
			isPrimary := att.Id == attachments[0].Id
			if att.GetBool("primary") != isPrimary {
				att.Set("primary", isPrimary)
				if err := s.app.Save(att); err != nil {
					return completed, err
				}
				changed = true
			}
		}
		if changed {
			completed++
		}
	}
	return completed, nil
}

func (s *Store) highestAssetID(groupID string) (int, error) {
	items, err := s.findAll(collections.Items, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return 0, err
	}
	highest := 0
	for _, item := range items {
		if v := item.GetInt("asset_id"); v > highest {
			highest = v
		}
	}
	return highest, nil
}

func (s *Store) NextAssetID(groupID string) (int, error) {
	highest, err := s.highestAssetID(groupID)
	if err != nil {
		return 0, err
	}
	return highest + 1, nil
}

func (s *Store) CreateItemWithAssetID(ctx context.Context, groupID string, data map[string]any, autoIncrement bool) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId(collections.Items)
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	for k, v := range data {
		rec.Set(k, v)
	}
	if autoIncrement {
		assetID, err := s.NextAssetID(groupID)
		if err != nil {
			return nil, err
		}
		rec.Set("asset_id", assetID)
	}
	if err := s.app.Save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}
