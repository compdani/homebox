package store

import (
	"context"
	"strings"
	"time"

	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func (s *Store) StatsGroup(ctx context.Context, groupID string) (GroupStatistics, error) {
	var stats GroupStatistics

	users, err := s.findAll(collections.Users, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return stats, err
	}
	stats.TotalUsers = len(users)

	items, err := s.findAll(collections.Items, "group = {:gid} && archived = false", dbx.Params{"gid": groupID})
	if err != nil {
		return stats, err
	}
	stats.TotalItems = len(items)
	for _, item := range items {
		stats.TotalItemPrice += item.GetFloat("purchase_price") * float64(item.GetInt("quantity"))
		if item.GetBool("lifetime_warranty") {
			stats.TotalWithWarranty++
			continue
		}
		if exp := item.GetDateTime("warranty_expires"); !exp.IsZero() && exp.Time().After(time.Now()) {
			stats.TotalWithWarranty++
		}
	}

	locations, err := s.findAll(collections.Locations, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return stats, err
	}
	stats.TotalLocations = len(locations)

	labels, err := s.findAll(collections.Labels, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return stats, err
	}
	stats.TotalLabels = len(labels)

	return stats, nil
}

func (s *Store) StatsLocationsByPurchasePrice(ctx context.Context, groupID string) ([]TotalsByOrganizer, error) {
	locations, err := s.findAll(collections.Locations, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	items, err := s.findAll(collections.Items, "group = {:gid} && archived = false", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	totals := make(map[string]float64)
	names := make(map[string]string)
	for _, loc := range locations {
		names[loc.Id] = loc.GetString("name")
	}

	for _, item := range items {
		locID := item.GetString("location")
		if locID == "" {
			continue
		}
		totals[locID] += item.GetFloat("purchase_price")
	}

	result := make([]TotalsByOrganizer, 0, len(totals))
	for id, total := range totals {
		result = append(result, TotalsByOrganizer{
			ID:    id,
			Name:  names[id],
			Total: total,
		})
	}
	return result, nil
}

func (s *Store) StatsLabelsByPurchasePrice(ctx context.Context, groupID string) ([]TotalsByOrganizer, error) {
	labels, err := s.findAll(collections.Labels, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	items, err := s.findAll(collections.Items, "group = {:gid} && archived = false", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	labelNames := make(map[string]string)
	for _, label := range labels {
		labelNames[label.Id] = label.GetString("name")
	}

	totals := make(map[string]float64)
	for _, item := range items {
		for _, labelID := range item.GetStringSlice("labels") {
			totals[labelID] += item.GetFloat("purchase_price")
		}
	}

	result := make([]TotalsByOrganizer, 0, len(totals))
	for id, total := range totals {
		result = append(result, TotalsByOrganizer{
			ID:    id,
			Name:  labelNames[id],
			Total: total,
		})
	}
	return result, nil
}

func (s *Store) StatsPurchasePrice(ctx context.Context, groupID string, start, end time.Time) (*ValueOverTime, error) {
	items, err := s.findAll(collections.Items, "group = {:gid} && archived = false", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	stats := &ValueOverTime{Start: start, End: end}
	for _, item := range items {
		created := item.GetDateTime("created").Time()
		price := item.GetFloat("purchase_price")
		if created.Before(start) {
			stats.PriceAtStart += price
		}
		if created.Before(end) || created.Equal(end) {
			stats.PriceAtEnd += price
		}
		if !created.Before(start) && !created.After(end) {
			stats.Entries = append(stats.Entries, ValueOverTimeEntry{
				Date:  created,
				Value: price,
				Name:  item.GetString("name"),
			})
		}
	}
	return stats, nil
}

func (s *Store) GroupByID(ctx context.Context, groupID string) (Group, error) {
	rec, err := s.app.FindRecordById(collections.Groups, groupID)
	if err != nil {
		return Group{}, err
	}
	return mapGroup(rec), nil
}

func (s *Store) GroupUpdate(ctx context.Context, groupID string, data GroupUpdate) (Group, error) {
	rec, err := s.app.FindRecordById(collections.Groups, groupID)
	if err != nil {
		return Group{}, err
	}
	rec.Set("name", data.Name)
	rec.Set("currency", strings.ToLower(data.Currency))
	if err := s.app.Save(rec); err != nil {
		return Group{}, err
	}
	return mapGroup(rec), nil
}

func mapGroup(rec *core.Record) Group {
	return Group{
		ID:        rec.Id,
		Name:      rec.GetString("name"),
		Currency:  strings.ToUpper(rec.GetString("currency")),
		CreatedAt: rec.GetDateTime("created").Time(),
		UpdatedAt: rec.GetDateTime("updated").Time(),
	}
}
