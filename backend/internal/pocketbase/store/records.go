package store

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func (s *Store) GetGroupRecord(collection, id, groupID string) (*core.Record, error) {
	return s.findOne(collection, "id = {:id} && group = {:gid}", dbx.Params{
		"id":  id,
		"gid": groupID,
	})
}
