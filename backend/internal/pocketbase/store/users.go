package store

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/hay-kot/homebox/backend/pkgs/hasher"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

var defaultLocationNames = []string{
	"Living Room", "Garage", "Kitchen", "Bedroom", "Bathroom", "Office", "Attic", "Basement",
}

var defaultLabelNames = []string{
	"Appliances", "IOT", "Electronics", "Servers", "General", "Important",
}

func (s *Store) RegisterUser(ctx context.Context, data UserRegistration) (*core.Record, error) {
	creatingGroup := data.GroupToken == ""
	var groupID string

	if creatingGroup {
		groupRec, err := s.createGroup("Home")
		if err != nil {
			return nil, err
		}
		groupID = groupRec.Id
	} else {
		inv, err := s.findInvitationByToken(data.GroupToken)
		if err != nil {
			return nil, err
		}
		if exp := inv.GetDateTime("expires_at"); !exp.IsZero() && exp.Time().Before(time.Now()) {
			return nil, errInvitationExpired
		}
		groupID = inv.GetString("group")
		if inv.GetInt("uses") <= 0 {
			return nil, errInvitationExpired
		}
		inv.Set("uses", inv.GetInt("uses")-1)
		if err := s.app.Save(inv); err != nil {
			return nil, err
		}
	}

	usersCollection, err := s.app.FindCollectionByNameOrId(collections.Users)
	if err != nil {
		return nil, err
	}

	user := core.NewRecord(usersCollection)
	user.Set("name", data.Name)
	user.SetEmail(data.Email)
	user.SetPassword(data.Password)
	user.Set("role", map[bool]string{true: "owner", false: "user"}[creatingGroup])
	user.Set("group", groupID)
	user.SetVerified(true)

	if err := s.app.Save(user); err != nil {
		return nil, err
	}

	if creatingGroup {
		if err := s.bootstrapGroupDefaults(groupID); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *Store) createGroup(name string) (*core.Record, error) {
	collection, err := s.app.FindCollectionByNameOrId(collections.Groups)
	if err != nil {
		return nil, err
	}
	rec := core.NewRecord(collection)
	rec.Set("name", name)
	rec.Set("currency", "usd")
	if err := s.app.Save(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *Store) bootstrapGroupDefaults(groupID string) error {
	for _, name := range defaultLabelNames {
		if err := s.createLabel(groupID, name); err != nil {
			return err
		}
	}
	for _, name := range defaultLocationNames {
		if err := s.createLocation(groupID, name); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) createLabel(groupID, name string) error {
	collection, err := s.app.FindCollectionByNameOrId(collections.Labels)
	if err != nil {
		return err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	rec.Set("name", name)
	return s.app.Save(rec)
}

func (s *Store) createLocation(groupID, name string) error {
	collection, err := s.app.FindCollectionByNameOrId(collections.Locations)
	if err != nil {
		return err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	rec.Set("name", name)
	return s.app.Save(rec)
}

func (s *Store) findInvitationByToken(raw string) (*core.Record, error) {
	hash := hasher.HashToken(raw)
	return s.findOne(collections.GroupInvitations, "token_hash = {:hash}", dbx.Params{"hash": hex.EncodeToString(hash)})
}

func (s *Store) CreateInvitation(ctx context.Context, groupID string, uses int, expiresAt time.Time) (GroupInvitation, string, error) {
	token := hasher.GenerateToken()
	collection, err := s.app.FindCollectionByNameOrId(collections.GroupInvitations)
	if err != nil {
		return GroupInvitation{}, "", err
	}
	rec := core.NewRecord(collection)
	rec.Set("group", groupID)
	rec.Set("token_hash", hex.EncodeToString(token.Hash))
	rec.Set("uses", uses)
	rec.Set("expires_at", expiresAt)
	if err := s.app.Save(rec); err != nil {
		return GroupInvitation{}, "", err
	}
	group, _ := s.GroupByID(ctx, groupID)
	return GroupInvitation{
		ID:        rec.Id,
		ExpiresAt: expiresAt,
		Uses:      uses,
		Token:     token.Raw,
		Group:     group,
	}, token.Raw, nil
}

func (s *Store) PurgeInvitations(ctx context.Context) (int, error) {
	records, err := s.findAll(collections.GroupInvitations, "uses <= 0 || expires_at < {:now}", dbx.Params{"now": time.Now()})
	if err != nil {
		return 0, err
	}
	for _, rec := range records {
		if err := s.app.Delete(rec); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

var (
	errInvitationNotFound = errStore("invitation not found")
	errInvitationExpired  = errStore("invitation expired")
)

type errStore string

func (e errStore) Error() string { return string(e) }
