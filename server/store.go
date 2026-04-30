package main

import (
	"errors"
	"time"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/larkox/mattermost-plugin-badges/server/sqlstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

var errInvalidBadge = errors.New("invalid badge")
var errBadgeNotFound = errors.New("badge not found")

type Store interface {
	GetUserBadges(userID string) ([]*badgesmodel.UserBadge, error)
	GetAllBadges() ([]*badgesmodel.AllBadgesBadge, error)
	GetBadgeDetails(badgeID badgesmodel.BadgeID) (*badgesmodel.BadgeDetails, error)

	GetRawBadges() ([]*badgesmodel.Badge, error)
	GetRawTypes() (badgesmodel.BadgeTypeList, error)

	AddBadge(badge *badgesmodel.Badge) (*badgesmodel.Badge, error)
	GrantBadge(badgeID badgesmodel.BadgeID, userID string, grantedBy string, reason string) (bool, error)
	AddType(t *badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error)
	GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error)
	GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error)
	UpdateType(t *badgesmodel.BadgeTypeDefinition) error
	UpdateBadge(b *badgesmodel.Badge) error
	DeleteType(tID badgesmodel.BadgeType) error
	DeleteBadge(bID badgesmodel.BadgeID) error

	AddSubscription(tID badgesmodel.BadgeType, cID string) error
	RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error
	GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error)
	GetChannelSubscriptions(cID string) ([]*badgesmodel.BadgeTypeDefinition, error)

	EnsureBadges(badges []*badgesmodel.Badge, pluginID, botID string) ([]*badgesmodel.Badge, error)
}

type sqlStoreAdapter struct {
	sql *sqlstore.SQLStore
	api plugin.API
}

func NewSQLStoreAdapter(sql *sqlstore.SQLStore, api plugin.API) Store {
	return &sqlStoreAdapter{sql: sql, api: api}
}

func (s *sqlStoreAdapter) GetUserBadges(userID string) ([]*badgesmodel.UserBadge, error) {
	ownership, err := s.sql.GetUserOwnership(userID)
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.UserBadge{}
	for _, o := range ownership {
		badge, err := s.sql.GetBadge(badgesmodel.BadgeID(o.BadgeID))
		if err != nil {
			continue
		}

		grantedByName := "unknown"
		u, appErr := s.api.GetUser(o.GrantedBy)
		if appErr == nil {
			grantedByName = s.getDisplayName(u)
		}

		typeName := "unknown"
		t, err := s.sql.GetType(badge.Type)
		if err == nil {
			typeName = t.Name
		}

		reason := ""
		if o.Reason != nil {
			reason = *o.Reason
		}

		out = append(out, &badgesmodel.UserBadge{
			Badge: *badge,
			Ownership: badgesmodel.Ownership{
				User:      o.UserID,
				Badge:     badgesmodel.BadgeID(o.BadgeID),
				Time:      time.UnixMilli(o.GrantedAt),
				Reason:    reason,
				GrantedBy: o.GrantedBy,
			},
			GrantedByUsername: grantedByName,
			TypeName:          typeName,
		})
	}

	return out, nil
}

func (s *sqlStoreAdapter) GetAllBadges() ([]*badgesmodel.AllBadgesBadge, error) {
	badges, err := s.sql.GetAllBadges()
	if err != nil {
		return nil, err
	}

	allOwnership, err := s.sql.GetAllOwnership()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.AllBadgesBadge{}
	for _, b := range badges {
		allBadge := &badgesmodel.AllBadgesBadge{Badge: *b}
		grantedTo := map[string]bool{}
		for _, o := range allOwnership {
			if o.BadgeID != string(b.ID) {
				continue
			}
			allBadge.GrantedTimes++
			if !grantedTo[o.UserID] {
				allBadge.Granted++
				grantedTo[o.UserID] = true
			}
		}

		allBadge.TypeName = "unknown"
		t, err := s.sql.GetType(b.Type)
		if err == nil {
			allBadge.TypeName = t.Name
		}
		out = append(out, allBadge)
	}

	return out, nil
}

func (s *sqlStoreAdapter) GetBadgeDetails(id badgesmodel.BadgeID) (*badgesmodel.BadgeDetails, error) {
	badge, err := s.sql.GetBadge(id)
	if err != nil {
		return nil, err
	}

	ownershipRows, err := s.sql.GetBadgeOwnership(id)
	if err != nil {
		return nil, err
	}

	owners := make(badgesmodel.OwnershipList, 0, len(ownershipRows))
	for _, o := range ownershipRows {
		reason := ""
		if o.Reason != nil {
			reason = *o.Reason
		}
		owners = append(owners, badgesmodel.Ownership{
			User:      o.UserID,
			Badge:     badgesmodel.BadgeID(o.BadgeID),
			Time:      time.UnixMilli(o.GrantedAt),
			Reason:    reason,
			GrantedBy: o.GrantedBy,
		})
	}

	createdByName := "unknown"
	u, appErr := s.api.GetUser(badge.CreatedBy)
	if appErr == nil {
		createdByName = s.getDisplayName(u)
	}

	typeName := "unknown"
	t, err := s.sql.GetType(badge.Type)
	if err == nil {
		typeName = t.Name
	}

	return &badgesmodel.BadgeDetails{
		Badge:             *badge,
		Owners:            owners,
		CreatedByUsername: createdByName,
		TypeName:          typeName,
	}, nil
}

func (s *sqlStoreAdapter) GetRawBadges() ([]*badgesmodel.Badge, error) {
	return s.sql.GetAllBadges()
}

func (s *sqlStoreAdapter) GetRawTypes() (badgesmodel.BadgeTypeList, error) {
	return s.sql.GetAllTypes()
}

func (s *sqlStoreAdapter) AddBadge(b *badgesmodel.Badge) (*badgesmodel.Badge, error) {
	if !b.IsValid() {
		return nil, errInvalidBadge
	}

	_, err := s.sql.GetType(b.Type)
	if err != nil {
		return nil, errors.New("missing badge type")
	}

	b.ID = badgesmodel.BadgeID(s.sql.NewBadgeID())
	if err := s.sql.InsertBadge(b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *sqlStoreAdapter) GrantBadge(badgeID badgesmodel.BadgeID, userID string, grantedBy string, reason string) (bool, error) {
	badge, err := s.sql.GetBadge(badgeID)
	if err != nil {
		return false, err
	}

	if !badge.Multiple {
		owned, err := s.sql.IsOwned(userID, badgeID)
		if err != nil {
			return false, err
		}
		if owned {
			return false, nil
		}
	}

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	row := &sqlstore.OwnershipRow{
		ID:        s.sql.NewOwnershipID(),
		UserID:    userID,
		BadgeID:   string(badgeID),
		GrantedBy: grantedBy,
		Reason:    reasonPtr,
		GrantedAt: model.GetMillis(),
	}

	if err := s.sql.InsertOwnership(row); err != nil {
		return false, err
	}

	return true, nil
}

func (s *sqlStoreAdapter) AddType(t *badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error) {
	t.ID = badgesmodel.BadgeType(s.sql.NewTypeID())
	if err := s.sql.InsertType(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *sqlStoreAdapter) GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error) {
	return s.sql.GetType(tID)
}

func (s *sqlStoreAdapter) GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	return s.sql.GetBadge(badgeID)
}

func (s *sqlStoreAdapter) UpdateType(t *badgesmodel.BadgeTypeDefinition) error {
	return s.sql.UpdateType(t)
}

func (s *sqlStoreAdapter) UpdateBadge(b *badgesmodel.Badge) error {
	return s.sql.UpdateBadge(b)
}

func (s *sqlStoreAdapter) DeleteType(tID badgesmodel.BadgeType) error {
	badges, err := s.sql.GetBadgesByType(tID)
	if err != nil {
		return err
	}

	for _, b := range badges {
		if err := s.DeleteBadge(b.ID); err != nil {
			return err
		}
	}

	if err := s.sql.DeleteSubscriptionsByType(tID); err != nil {
		return err
	}

	return s.sql.DeleteType(tID)
}

func (s *sqlStoreAdapter) DeleteBadge(bID badgesmodel.BadgeID) error {
	if err := s.sql.DeleteBadgeOwnership(bID); err != nil {
		return err
	}
	return s.sql.DeleteBadge(bID)
}

func (s *sqlStoreAdapter) AddSubscription(tID badgesmodel.BadgeType, cID string) error {
	return s.sql.InsertSubscription(tID, cID)
}

func (s *sqlStoreAdapter) RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error {
	return s.sql.DeleteSubscription(tID, cID)
}

func (s *sqlStoreAdapter) GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error) {
	return s.sql.GetTypeSubscriptionChannels(tID)
}

func (s *sqlStoreAdapter) GetChannelSubscriptions(cID string) ([]*badgesmodel.BadgeTypeDefinition, error) {
	typeIDs, err := s.sql.GetChannelSubscriptionTypes(cID)
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.BadgeTypeDefinition{}
	for _, tID := range typeIDs {
		t, err := s.sql.GetType(tID)
		if err != nil {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func (s *sqlStoreAdapter) EnsureBadges(badges []*badgesmodel.Badge, pluginID, botID string) ([]*badgesmodel.Badge, error) {
	tDef, err := s.sql.GetTypeByCreator(botID)
	if err != nil {
		return nil, err
	}

	if tDef == nil {
		newType := &badgesmodel.BadgeTypeDefinition{
			ID:        badgesmodel.BadgeType(s.sql.NewTypeID()),
			Name:      "Plugin badges: " + pluginID,
			CreatedBy: botID,
		}
		if err := s.sql.InsertType(newType); err != nil {
			return nil, err
		}
		tDef = newType
	}

	existingBadges, err := s.sql.GetBadgesByCreator(botID)
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.Badge{}
	for _, pb := range badges {
		found := false
		for _, b := range existingBadges {
			if b.Name == pb.Name {
				found = true
				out = append(out, b)
				break
			}
		}
		if !found {
			pb.Type = tDef.ID
			pb.CreatedBy = botID
			pb.ID = badgesmodel.BadgeID(s.sql.NewBadgeID())
			if err := s.sql.InsertBadge(pb); err != nil {
				return nil, err
			}
			out = append(out, pb)
		}
	}

	return out, nil
}

func (s *sqlStoreAdapter) getDisplayName(u *model.User) string {
	conf := s.api.GetConfig()
	if conf != nil && conf.TeamSettings.TeammateNameDisplay != nil {
		return u.GetDisplayName(*conf.TeamSettings.TeammateNameDisplay)
	}
	return u.Username
}
