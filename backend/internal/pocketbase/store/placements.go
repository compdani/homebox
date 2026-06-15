package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type PlaceItemRequest struct {
	ProductID  string `json:"productId"`
	ItemID     string `json:"itemId"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
}

type PlaceItemResult struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	Created  bool   `json:"created"`
}

func (s *Store) PlaceItem(ctx context.Context, groupID string, req PlaceItemRequest) (*PlaceItemResult, error) {
	if req.LocationID == "" {
		return nil, fmt.Errorf("locationId is required")
	}
	if req.Quantity < 1 {
		return nil, fmt.Errorf("quantity must be at least 1")
	}

	location, err := s.findOne(collections.Locations, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  req.LocationID,
		"gid": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("location not found")
	}
	_ = location

	if req.ProductID != "" {
		return s.placeProduct(ctx, groupID, req)
	}
	if req.ItemID != "" {
		return s.placeUniqueItem(ctx, groupID, req)
	}
	return nil, fmt.Errorf("productId or itemId is required")
}

func (s *Store) placeProduct(ctx context.Context, groupID string, req PlaceItemRequest) (*PlaceItemResult, error) {
	product, err := s.findOne(collections.Products, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  req.ProductID,
		"gid": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	existing, err := s.findOne(collections.Items,
		"group = {:gid} && product = {:pid} && location = {:lid}",
		dbx.Params{"gid": groupID, "pid": req.ProductID, "lid": req.LocationID},
	)
	if err == nil {
		newQty := existing.GetInt("quantity") + req.Quantity
		if newQty < 1 {
			newQty = req.Quantity
		}
		existing.Set("quantity", newQty)
		if err := s.app.Save(existing); err != nil {
			return nil, err
		}
		return &PlaceItemResult{ID: existing.Id, Quantity: newQty, Created: false}, nil
	}
	if !errors.Is(err, errNotFound) {
		return nil, err
	}

	itemCollection, err := s.app.FindCollectionByNameOrId(collections.Items)
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(itemCollection)
	rec.Set("group", groupID)
	rec.Set("product", req.ProductID)
	rec.Set("location", req.LocationID)
	rec.Set("name", product.GetString("name"))
	rec.Set("quantity", req.Quantity)
	if err := s.app.Save(rec); err != nil {
		return nil, err
	}
	return &PlaceItemResult{ID: rec.Id, Quantity: req.Quantity, Created: true}, nil
}

func (s *Store) placeUniqueItem(ctx context.Context, groupID string, req PlaceItemRequest) (*PlaceItemResult, error) {
	item, err := s.findOne(collections.Items, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  req.ItemID,
		"gid": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}
	if item.GetString("product") != "" {
		return nil, fmt.Errorf("product-linked items cannot be placed via item scan")
	}

	item.Set("location", req.LocationID)
	item.Set("quantity", req.Quantity)
	if err := s.app.Save(item); err != nil {
		return nil, err
	}
	return &PlaceItemResult{ID: item.Id, Quantity: req.Quantity, Created: false}, nil
}

type UnplaceItemRequest struct {
	ProductID  string `json:"productId"`
	LocationID string `json:"locationId"`
	Quantity   int    `json:"quantity"`
}

type UnplaceItemResult struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	Removed  bool   `json:"removed"`
}

func (s *Store) UnplaceProduct(ctx context.Context, groupID string, req UnplaceItemRequest) (*UnplaceItemResult, error) {
	if req.ProductID == "" {
		return nil, fmt.Errorf("productId is required")
	}
	if req.LocationID == "" {
		return nil, fmt.Errorf("locationId is required")
	}
	if req.Quantity < 1 {
		return nil, fmt.Errorf("quantity must be at least 1")
	}

	_, err := s.findOne(collections.Products, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  req.ProductID,
		"gid": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	_, err = s.findOne(collections.Locations, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  req.LocationID,
		"gid": groupID,
	})
	if err != nil {
		return nil, fmt.Errorf("location not found")
	}

	existing, err := s.findOne(collections.Items,
		"group = {:gid} && product = {:pid} && location = {:lid}",
		dbx.Params{"gid": groupID, "pid": req.ProductID, "lid": req.LocationID},
	)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil, fmt.Errorf("product not at this location")
		}
		return nil, err
	}

	currentQty := existing.GetInt("quantity")
	if currentQty < req.Quantity {
		return nil, fmt.Errorf("cannot remove %d; only %d at this location", req.Quantity, currentQty)
	}

	newQty := currentQty - req.Quantity
	if newQty <= 0 {
		if err := s.app.Delete(existing); err != nil {
			return nil, err
		}
		return &UnplaceItemResult{ID: existing.Id, Quantity: 0, Removed: true}, nil
	}

	existing.Set("quantity", newQty)
	if err := s.app.Save(existing); err != nil {
		return nil, err
	}
	return &UnplaceItemResult{ID: existing.Id, Quantity: newQty, Removed: false}, nil
}
