package sqlstore

import (
	"encoding/json"
	"time"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	kvMigrationDoneKey = "KVMigrationDone"
	kvMigrationDoneVal = "true"

	kvKeyBadges        = "badges"
	kvKeyOwnership     = "ownership"
	kvKeyTypes         = "types"
	kvKeySubscriptions = "subs"
)

type kvOwnership struct {
	User      string    `json:"user"`
	GrantedBy string    `json:"granted_by"`
	Badge     string    `json:"badge"`
	Reason    string    `json:"reason"`
	Time      time.Time `json:"time"`
}

type kvSubscription struct {
	TypeID    string `json:"TypeID"`
	ChannelID string `json:"ChannelID"`
}

// MigrateFromKV reads all KV entries and inserts them into SQL tables.
// Idempotent: skips if migration was already completed.
// Existing KV data is NOT deleted (safe rollback).
func (s *SQLStore) MigrateFromKV(api plugin.API) (retErr error) {
	done, err := s.getSystemValue(s.db, kvMigrationDoneKey)
	if err != nil {
		logrus.WithError(err).Warn("Could not check KV migration flag, proceeding with migration")
	}
	if done == kvMigrationDoneVal {
		logrus.Info("KV->DB migration already completed, skipping")
		return nil
	}

	logrus.Info("Starting KV->DB migration for Badges plugin")

	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "could not begin migration transaction")
	}
	defer s.finalizeTransaction(tx, &retErr)

	migratedTypes, err := s.migrateTypes(api, tx)
	if err != nil {
		return errors.Wrap(err, "failed to migrate types")
	}

	migratedBadges, err := s.migrateBadges(api, tx)
	if err != nil {
		return errors.Wrap(err, "failed to migrate badges")
	}

	migratedOwnership, err := s.migrateOwnership(api, tx)
	if err != nil {
		return errors.Wrap(err, "failed to migrate ownership")
	}

	migratedSubs, err := s.migrateSubscriptions(api, tx)
	if err != nil {
		return errors.Wrap(err, "failed to migrate subscriptions")
	}

	if err := s.setSystemValue(tx, kvMigrationDoneKey, kvMigrationDoneVal); err != nil {
		return errors.Wrap(err, "failed to set KV migration done flag")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "could not commit migration transaction")
	}

	logrus.WithFields(logrus.Fields{
		"types":         migratedTypes,
		"badges":        migratedBadges,
		"ownership":     migratedOwnership,
		"subscriptions": migratedSubs,
	}).Info("KV->DB migration completed")

	return nil
}

func (s *SQLStore) migrateTypes(api plugin.API, tx execer) (int, error) {
	data, appErr := api.KVGet(kvKeyTypes)
	if appErr != nil {
		return 0, errors.New(appErr.Error())
	}
	if data == nil {
		return 0, nil
	}

	var types []*badgesmodel.BadgeTypeDefinition
	if err := json.Unmarshal(data, &types); err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal KV types")
	}

	count := 0
	for _, t := range types {
		canGrant, err := marshalPermissionScheme(t.CanGrant)
		if err != nil {
			logrus.WithError(err).WithField("type_id", t.ID).Warn("Skipping type with bad can_grant")
			continue
		}
		canCreate, err := marshalPermissionScheme(t.CanCreate)
		if err != nil {
			logrus.WithError(err).WithField("type_id", t.ID).Warn("Skipping type with bad can_create")
			continue
		}

		_, err = tx.Exec(
			`INSERT INTO badge_types (id, name, frame, created_by, can_grant, can_create)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (id) DO NOTHING`,
			string(t.ID), t.Name, t.Frame, t.CreatedBy, canGrant, canCreate,
		)
		if err != nil {
			logrus.WithError(err).WithField("type_id", t.ID).Warn("Failed to insert migrated type")
			continue
		}
		count++
	}
	return count, nil
}

func (s *SQLStore) migrateBadges(api plugin.API, tx execer) (int, error) {
	data, appErr := api.KVGet(kvKeyBadges)
	if appErr != nil {
		return 0, errors.New(appErr.Error())
	}
	if data == nil {
		return 0, nil
	}

	var badges []*badgesmodel.Badge
	if err := json.Unmarshal(data, &badges); err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal KV badges")
	}

	count := 0
	for _, b := range badges {
		_, err := tx.Exec(
			`INSERT INTO badges (id, name, description, image, image_type, multiple, type_id, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (id) DO NOTHING`,
			string(b.ID), b.Name, b.Description, b.Image,
			string(b.ImageType), b.Multiple, string(b.Type), b.CreatedBy,
		)
		if err != nil {
			logrus.WithError(err).WithField("badge_id", b.ID).Warn("Failed to insert migrated badge")
			continue
		}
		count++
	}
	return count, nil
}

func (s *SQLStore) migrateOwnership(api plugin.API, tx execer) (int, error) {
	data, appErr := api.KVGet(kvKeyOwnership)
	if appErr != nil {
		return 0, errors.New(appErr.Error())
	}
	if data == nil {
		return 0, nil
	}

	var ownership []kvOwnership
	if err := json.Unmarshal(data, &ownership); err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal KV ownership")
	}

	count := 0
	for _, o := range ownership {
		id := model.NewId()
		grantedAtMs := o.Time.UnixMilli()

		var reason *string
		if o.Reason != "" {
			reason = &o.Reason
		}

		_, err := tx.Exec(
			`INSERT INTO badge_ownership (id, user_id, badge_id, granted_by, reason, granted_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			id, o.User, o.Badge, o.GrantedBy, reason, grantedAtMs,
		)
		if err != nil {
			logrus.WithError(err).WithField("user", o.User).WithField("badge", o.Badge).Warn("Failed to insert migrated ownership")
			continue
		}
		count++
	}
	return count, nil
}

func (s *SQLStore) migrateSubscriptions(api plugin.API, tx execer) (int, error) {
	data, appErr := api.KVGet(kvKeySubscriptions)
	if appErr != nil {
		return 0, errors.New(appErr.Error())
	}
	if data == nil {
		return 0, nil
	}

	var subs []kvSubscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal KV subscriptions")
	}

	count := 0
	for _, sub := range subs {
		_, err := tx.Exec(
			`INSERT INTO badge_subscriptions (type_id, channel_id)
			 VALUES ($1, $2)
			 ON CONFLICT (type_id, channel_id) DO NOTHING`,
			sub.TypeID, sub.ChannelID,
		)
		if err != nil {
			logrus.WithError(err).WithField("type_id", sub.TypeID).Warn("Failed to insert migrated subscription")
			continue
		}
		count++
	}
	return count, nil
}
