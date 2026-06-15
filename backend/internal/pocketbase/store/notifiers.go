package store

import (
	"context"
	"strings"
	"time"

	"github.com/containrrr/shoutrrr"
	"github.com/hay-kot/homebox/backend/internal/pocketbase/collections"
	"github.com/pocketbase/dbx"
	"github.com/rs/zerolog/log"
)

func (s *Store) SendNotifiersToday(ctx context.Context) error {
	groups, err := s.findAll(collections.Groups, "", dbx.Params{})
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")

	for _, group := range groups {
		entries, err := s.findAll(
			collections.Maintenance,
			"group = {:gid} && scheduled_date ~ {:today}",
			dbx.Params{"gid": group.Id, "today": today},
		)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			log.Debug().Str("group_name", group.GetString("name")).Str("group_id", group.Id).Msg("No scheduled maintenance for today")
			continue
		}

		notifiers, err := s.findAll(
			collections.Notifiers,
			"group = {:gid} && is_active = true",
			dbx.Params{"gid": group.Id},
		)
		if err != nil {
			return err
		}

		var bldr strings.Builder
		bldr.WriteString("Homebox Maintenance for (")
		bldr.WriteString(today)
		bldr.WriteString("):\n")
		for _, entry := range entries {
			bldr.WriteString(" - ")
			bldr.WriteString(entry.GetString("name"))
			bldr.WriteString("\n")
		}

		for _, notifier := range notifiers {
			if err := shoutrrr.Send(notifier.GetString("url"), bldr.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) TestNotifier(ctx context.Context, userID, notifierID string) error {
	rec, err := s.app.FindRecordById(collections.Notifiers, notifierID)
	if err != nil {
		return err
	}
	if rec.GetString("user") != userID {
		return errNotFound
	}
	return shoutrrr.Send(rec.GetString("url"), "Homebox test notification")
}
