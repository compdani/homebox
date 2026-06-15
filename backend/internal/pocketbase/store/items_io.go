package store

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/hay-kot/homebox/backend/internal/core/services/reporting"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type BillOfMaterialsEntry struct {
	PurchaseDate string  `csv:"Purchase Date"`
	Name         string  `csv:"Name"`
	Description  string  `csv:"Description"`
	Manufacturer string  `csv:"Manufacturer"`
	SerialNumber string  `csv:"Serial Number"`
	ModelNumber  string  `csv:"Model Number"`
	Quantity     int     `csv:"Quantity"`
	Price        float64 `csv:"Price"`
	TotalPrice   float64 `csv:"Total Price"`
}

func (s *Store) ExportBillOfMaterials(ctx context.Context, groupID string) ([]byte, error) {
	items, err := s.findAll(collections.Items, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}
	entries := make([]BillOfMaterialsEntry, len(items))
	for i, item := range items {
		qty := item.GetInt("quantity")
		if qty == 0 {
			qty = 1
		}
		price := item.GetFloat("purchase_price")
		entries[i] = BillOfMaterialsEntry{
			PurchaseDate: formatDate(item.GetDateTime("purchase_time")),
			Name:         item.GetString("name"),
			Description:  item.GetString("description"),
			Manufacturer: item.GetString("manufacturer"),
			SerialNumber: item.GetString("serial_number"),
			ModelNumber:  item.GetString("model_number"),
			Quantity:     qty,
			Price:        price,
			TotalPrice:   price * float64(qty),
		}
	}
	return gocsv.MarshalBytes(&entries)
}

func (s *Store) ExportCSV(ctx context.Context, groupID string) ([][]string, error) {
	items, err := s.listItemOuts(groupID)
	if err != nil {
		return nil, err
	}
	sheet := reporting.IOSheet{}
	for _, item := range items {
		row := reporting.ExportTSVRow{
			ImportRef: item.ImportRef,
			Name:      item.Name,
			Product:   item.Product,
		}
		sheet.Rows = append(sheet.Rows, row)
	}
	return sheet.TSV()
}

func (s *Store) ImportCSV(ctx context.Context, groupID string, data io.Reader, autoIncrementAssetID bool) (int, error) {
	sheet := reporting.IOSheet{}
	if err := sheet.Read(data); err != nil {
		return 0, err
	}

	labelMap, err := s.labelNameMap(groupID)
	if err != nil {
		return 0, err
	}
	productMap, err := s.productNameMap(groupID)
	if err != nil {
		return 0, err
	}
	locationMap, err := s.locationPathMap(groupID)
	if err != nil {
		return 0, err
	}

	highestAID := 0
	if autoIncrementAssetID {
		highestAID, err = s.highestAssetID(groupID)
		if err != nil {
			return 0, err
		}
	}

	finished := 0
	for _, row := range sheet.Rows {
		if row.ImportRef != "" {
			existing, _ := s.findOne(collections.Items, "group = {:gid} && import_ref = {:ref}", dbx.Params{"gid": groupID, "ref": row.ImportRef})
			if existing != nil {
				continue
			}
		}

		labelIDs := make([]string, 0, len(row.LabelStr))
		for _, label := range row.LabelStr {
			id, ok := labelMap[label]
			if !ok {
				rec, err := s.createLabelRecord(groupID, label)
				if err != nil {
					return finished, err
				}
				id = rec.Id
				labelMap[label] = id
			}
			labelIDs = append(labelIDs, id)
		}

		locationID, err := s.ensureLocationPath(groupID, row.Location, locationMap)
		if err != nil {
			return finished, err
		}

		productID := ""
		if row.Product != "" {
			productID, err = s.ensureProduct(groupID, row.Product, row.Manufacturer, row.ModelNumber, productMap)
			if err != nil {
				return finished, err
			}
		}

		if productID != "" && locationID != "" {
			existing, findErr := s.findOne(collections.Items,
				"group = {:gid} && product = {:pid} && location = {:lid}",
				dbx.Params{"gid": groupID, "pid": productID, "lid": locationID},
			)
			if findErr == nil {
				qty := row.Quantity
				if qty < 1 {
					qty = 1
				}
				existing.Set("quantity", existing.GetInt("quantity")+qty)
				if err := s.app.Save(existing); err != nil {
					return finished, err
				}
				finished++
				continue
			}
		}

		assetID := 0
		if !row.AssetID.Nil() {
			assetID = int(row.AssetID)
		}
		if autoIncrementAssetID && assetID == 0 {
			highestAID++
			assetID = highestAID
		}

		collection, _ := s.app.FindCollectionByNameOrId(collections.Items)
		rec := core.NewRecord(collection)
		rec.Set("group", groupID)
		if productID != "" {
			rec.Set("product", productID)
			if row.Name == "" {
				rec.Set("name", row.Product)
			} else {
				rec.Set("name", row.Name)
			}
		} else {
			rec.Set("name", row.Name)
		}
		rec.Set("description", row.Description)
		rec.Set("import_ref", row.ImportRef)
		rec.Set("asset_id", assetID)
		rec.Set("location", locationID)
		rec.Set("labels", labelIDs)
		rec.Set("quantity", row.Quantity)
		rec.Set("insured", row.Insured)
		rec.Set("archived", row.Archived)
		rec.Set("purchase_price", row.PurchasePrice)
		rec.Set("purchase_from", row.PurchaseFrom)
		rec.Set("manufacturer", row.Manufacturer)
		rec.Set("model_number", row.ModelNumber)
		rec.Set("serial_number", row.SerialNumber)
		rec.Set("lifetime_warranty", row.LifetimeWarranty)
		rec.Set("warranty_details", row.WarrantyDetails)
		rec.Set("sold_to", row.SoldTo)
		rec.Set("sold_price", row.SoldPrice)
		rec.Set("sold_notes", row.SoldNotes)
		rec.Set("notes", row.Notes)
		if err := s.app.Save(rec); err != nil {
			return finished, err
		}
		finished++
	}
	return finished, nil
}

type itemOutLite struct {
	ImportRef string
	Name      string
	Product   string
}

func (s *Store) listItemOuts(groupID string) ([]itemOutLite, error) {
	records, err := s.findAll(collections.Items, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}
	productNames := map[string]string{}
	products, err := s.findAll(collections.Products, "group = {:gid}", dbx.Params{"gid": groupID})
	if err == nil {
		for _, p := range products {
			productNames[p.Id] = p.GetString("name")
		}
	}
	out := make([]itemOutLite, len(records))
	for i, rec := range records {
		out[i] = itemOutLite{
			ImportRef: rec.GetString("import_ref"),
			Name:      rec.GetString("name"),
			Product:   productNames[rec.GetString("product")],
		}
	}
	return out, nil
}

func (s *Store) productNameMap(groupID string) (map[string]string, error) {
	records, err := s.findAll(collections.Products, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(records))
	for _, rec := range records {
		m[rec.GetString("name")] = rec.Id
	}
	return m, nil
}

func (s *Store) ensureProduct(groupID, name, manufacturer, modelNumber string, productMap map[string]string) (string, error) {
	if id, ok := productMap[name]; ok {
		return id, nil
	}
	collection, err := s.app.FindCollectionByNameOrId(collections.Products)
	if err != nil {
		return "", err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	rec.Set("name", name)
	rec.Set("manufacturer", manufacturer)
	rec.Set("model_number", modelNumber)
	if err := s.app.Save(rec); err != nil {
		return "", err
	}
	productMap[name] = rec.Id
	return rec.Id, nil
}

func (s *Store) labelNameMap(groupID string) (map[string]string, error) {
	records, err := s.findAll(collections.Labels, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(records))
	for _, rec := range records {
		m[rec.GetString("name")] = rec.Id
	}
	return m, nil
}

func (s *Store) locationPathMap(groupID string) (map[string]string, error) {
	records, err := s.findAll(collections.Locations, "group = {:gid}", dbx.Params{"gid": groupID})
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*core.Record, len(records))
	for _, rec := range records {
		byID[rec.Id] = rec
	}
	m := make(map[string]string, len(records))
	for _, rec := range records {
		path := s.locationFullPath(rec, byID)
		m[path] = rec.Id
	}
	return m, nil
}

func (s *Store) locationFullPath(rec *core.Record, byID map[string]*core.Record) string {
	parts := []string{rec.GetString("name")}
	current := rec
	for {
		parentID := current.GetString("parent")
		if parentID == "" {
			break
		}
		parent := byID[parentID]
		if parent == nil {
			break
		}
		parts = append([]string{parent.GetString("name")}, parts...)
		current = parent
	}
	return strings.Join(parts, "/")
}

func (s *Store) createLabelRecord(groupID, name string) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId(collections.Labels)
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	rec.Set("name", name)
	return rec, s.app.Save(rec)
}

func (s *Store) ensureLocationPath(groupID string, parts []string, locationMap map[string]string) (string, error) {
	if len(parts) == 0 {
		return "", nil
	}
	paths := []string{}
	var parentID string
	for _, part := range parts {
		paths = append(paths, part)
		path := strings.Join(paths, "/")
		if id, ok := locationMap[path]; ok {
			parentID = id
			continue
		}
		collection, err := s.app.FindCollectionByNameOrId(collections.Locations)
		if err != nil {
			return "", err
		}
		rec := core.NewRecord(collection)
		rec.Set("group", groupID)
		rec.Set("name", part)
		if parentID != "" {
			rec.Set("parent", parentID)
		}
		if err := s.app.Save(rec); err != nil {
			return "", err
		}
		parentID = rec.Id
		locationMap[path] = rec.Id
	}
	return locationMap[strings.Join(parts, "/")], nil
}

func formatDate(dt typesDateTime) string {
	if dt.IsZero() {
		return ""
	}
	return dt.Time().Format("2006-01-02")
}

type typesDateTime interface {
	Time() time.Time
	IsZero() bool
}
