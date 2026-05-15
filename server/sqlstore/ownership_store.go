package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type OwnershipRow struct {
	ID        string  `db:"id"`
	UserID    string  `db:"user_id"`
	BadgeID   string  `db:"badge_id"`
	GrantedBy string  `db:"granted_by"`
	Reason    *string `db:"reason"`
	GrantedAt int64   `db:"granted_at"`
}

func (s *SQLStore) InsertOwnership(o *OwnershipRow) error {
	_, err := s.execBuilder(s.db,
		sq.Insert("badge_ownership").
			Columns("id", "user_id", "badge_id", "granted_by", "reason", "granted_at").
			Values(o.ID, o.UserID, o.BadgeID, o.GrantedBy, o.Reason, o.GrantedAt),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert ownership %s", o.ID)
	}
	return nil
}

func (s *SQLStore) IsOwned(userID string, badgeID badgesmodel.BadgeID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM badge_ownership WHERE user_id = $1 AND badge_id = $2`
	if err := s.db.Get(&count, query, userID, string(badgeID)); err != nil {
		return false, errors.Wrap(err, "failed to check ownership")
	}
	return count > 0, nil
}

func (s *SQLStore) GetUserOwnership(userID string) ([]*OwnershipRow, error) {
	var rows []*OwnershipRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select("id", "user_id", "badge_id", "granted_by", "reason", "granted_at").
			From("badge_ownership").
			Where(sq.Eq{"user_id": userID}).
			OrderBy("granted_at DESC"),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ownership for user %s", userID)
	}
	if rows == nil {
		rows = []*OwnershipRow{}
	}
	return rows, nil
}

func (s *SQLStore) GetBadgeOwnership(badgeID badgesmodel.BadgeID) ([]*OwnershipRow, error) {
	var rows []*OwnershipRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select("id", "user_id", "badge_id", "granted_by", "reason", "granted_at").
			From("badge_ownership").
			Where(sq.Eq{"badge_id": string(badgeID)}).
			OrderBy("granted_at DESC"),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ownership for badge %s", badgeID)
	}
	if rows == nil {
		rows = []*OwnershipRow{}
	}
	return rows, nil
}

func (s *SQLStore) GetAllOwnership() ([]*OwnershipRow, error) {
	var rows []*OwnershipRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select("id", "user_id", "badge_id", "granted_by", "reason", "granted_at").
			From("badge_ownership").
			OrderBy("granted_at DESC"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all ownership")
	}
	if rows == nil {
		rows = []*OwnershipRow{}
	}
	return rows, nil
}

func (s *SQLStore) DeleteBadgeOwnership(badgeID badgesmodel.BadgeID) error {
	_, err := s.execBuilder(s.db,
		sq.Delete("badge_ownership").
			Where(sq.Eq{"badge_id": string(badgeID)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete ownership for badge %s", badgeID)
	}
	return nil
}

func (s *SQLStore) NewOwnershipID() string {
	return model.NewId()
}

// GetBadgeCountByUser is used by the inter-plugin API (/papi/v1/badge-count).
func (s *SQLStore) GetBadgeCountByUser(userID string, startMs, endMs int64) (int, error) {
	query := `
		SELECT COALESCE(COUNT(*), 0)
		FROM badge_ownership
		WHERE user_id = $1 AND granted_at >= $2 AND granted_at <= $3
	`
	var count int
	if err := s.db.Get(&count, query, userID, startMs, endMs); err != nil {
		return 0, errors.Wrapf(err, "failed to get badge count for user %s", userID)
	}
	return count, nil
}

