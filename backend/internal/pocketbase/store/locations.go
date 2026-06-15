package store

import (
	"context"

	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
)

func (s *Store) LocationTree(ctx context.Context, groupID string) ([]LocationTree, error) {
	records, err := s.findAll(collections.Locations, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}

	byID := make(map[string]LocationTree, len(records))
	children := make(map[string][]string)
	for _, rec := range records {
		byID[rec.Id] = LocationTree{
			ID:       rec.Id,
			Name:     rec.GetString("name"),
			Type:     "location",
			Children: []LocationTree{},
		}
		parent := rec.GetString("parent")
		if parent != "" {
			children[parent] = append(children[parent], rec.Id)
		}
	}

	var build func(id string) LocationTree
	build = func(id string) LocationTree {
		node := byID[id]
		for _, childID := range children[id] {
			node.Children = append(node.Children, build(childID))
		}
		return node
	}

	roots := make([]LocationTree, 0)
	for _, rec := range records {
		if rec.GetString("parent") == "" {
			roots = append(roots, build(rec.Id))
		}
	}
	return roots, nil
}

func (s *Store) ItemPath(ctx context.Context, groupID, itemID string) ([]ItemPath, error) {
	item, err := s.app.FindRecordById(collections.Items, itemID)
	if err != nil {
		return nil, err
	}
	if item.GetString("group") != groupID {
		return nil, errNotFound
	}

	path := []ItemPath{{ID: item.Id, Name: item.GetString("name")}}
	current := item
	for {
		parentID := current.GetString("parent")
		if parentID == "" {
			break
		}
		parent, err := s.app.FindRecordById(collections.Items, parentID)
		if err != nil {
			break
		}
		path = append([]ItemPath{{ID: parent.Id, Name: parent.GetString("name")}}, path...)
		current = parent
	}
	return path, nil
}

var errNotFound = errStore("not found")
