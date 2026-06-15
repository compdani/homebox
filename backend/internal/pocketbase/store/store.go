package store

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// Store provides data access for custom Homebox API routes via PocketBase.
type Store struct {
	app core.App
}

func New(app core.App) *Store {
	return &Store{app: app}
}

func (s *Store) App() core.App {
	return s.app
}

func (s *Store) findAll(collection, filter string, params dbx.Params) ([]*core.Record, error) {
	return s.app.FindRecordsByFilter(collection, filter, "", -1, 0, params)
}

func (s *Store) findOne(collection, filter string, params dbx.Params) (*core.Record, error) {
	records, err := s.app.FindRecordsByFilter(collection, filter, "", 1, 0, params)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errNotFound
	}
	return records[0], nil
}

func authGroupID(auth *core.Record) string {
	if auth == nil {
		return ""
	}
	return auth.GetString("group")
}
