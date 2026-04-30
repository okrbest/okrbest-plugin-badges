package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/pkg/errors"
)

func (s *SQLStore) InsertSubscription(typeID badgesmodel.BadgeType, channelID string) error {
	_, err := s.execBuilder(s.db,
		sq.Insert("badge_subscriptions").
			Columns("type_id", "channel_id").
			Values(string(typeID), channelID).
			Suffix("ON CONFLICT (type_id, channel_id) DO NOTHING"),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert subscription type=%s channel=%s", typeID, channelID)
	}
	return nil
}

func (s *SQLStore) DeleteSubscription(typeID badgesmodel.BadgeType, channelID string) error {
	_, err := s.execBuilder(s.db,
		sq.Delete("badge_subscriptions").
			Where(sq.Eq{"type_id": string(typeID), "channel_id": channelID}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete subscription type=%s channel=%s", typeID, channelID)
	}
	return nil
}

func (s *SQLStore) GetTypeSubscriptionChannels(typeID badgesmodel.BadgeType) ([]string, error) {
	var channelIDs []string
	err := s.selectBuilder(s.db, &channelIDs,
		sq.Select("channel_id").
			From("badge_subscriptions").
			Where(sq.Eq{"type_id": string(typeID)}),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get subscriptions for type %s", typeID)
	}
	if channelIDs == nil {
		channelIDs = []string{}
	}
	return channelIDs, nil
}

func (s *SQLStore) GetChannelSubscriptionTypes(channelID string) ([]badgesmodel.BadgeType, error) {
	var typeIDs []string
	err := s.selectBuilder(s.db, &typeIDs,
		sq.Select("type_id").
			From("badge_subscriptions").
			Where(sq.Eq{"channel_id": channelID}),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get subscriptions for channel %s", channelID)
	}

	result := make([]badgesmodel.BadgeType, 0, len(typeIDs))
	for _, id := range typeIDs {
		result = append(result, badgesmodel.BadgeType(id))
	}
	return result, nil
}

func (s *SQLStore) DeleteSubscriptionsByType(typeID badgesmodel.BadgeType) error {
	_, err := s.execBuilder(s.db,
		sq.Delete("badge_subscriptions").
			Where(sq.Eq{"type_id": string(typeID)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete subscriptions for type %s", typeID)
	}
	return nil
}
