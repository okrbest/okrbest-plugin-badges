package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/lib/pq"
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

// --- Dashboard aggregation queries ---

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

type GrantedBadgeRow struct {
	BadgeID   string `db:"badge_id" json:"badge_id"`
	BadgeName string `db:"badge_name" json:"badge_name"`
	GrantedBy string `db:"granted_by" json:"granted_by"`
	GrantedAt int64  `db:"granted_at" json:"granted_at"`
}

func (s *SQLStore) GetGrantedBadges(userID string, startMs, endMs int64) ([]*GrantedBadgeRow, error) {
	query := `
		SELECT o.badge_id, b.name AS badge_name, o.granted_by, o.granted_at
		FROM badge_ownership o
		JOIN badges b ON b.id = o.badge_id
		WHERE o.user_id = $1 AND o.granted_at >= $2 AND o.granted_at <= $3
		ORDER BY o.granted_at DESC
	`
	var rows []*GrantedBadgeRow
	if err := s.db.Select(&rows, query, userID, startMs, endMs); err != nil {
		return nil, errors.Wrapf(err, "failed to get granted badges for user %s", userID)
	}
	if rows == nil {
		rows = []*GrantedBadgeRow{}
	}
	return rows, nil
}

type MemberBadgeRow struct {
	UserID    string `db:"user_id" json:"user_id"`
	BadgeID   string `db:"badge_id" json:"badge_id"`
	BadgeName string `db:"badge_name" json:"badge_name"`
	GrantedAt int64  `db:"granted_at" json:"granted_at"`
}

func (s *SQLStore) GetRecentMemberBadges(userIDs []string, limit int) ([]*MemberBadgeRow, error) {
	if len(userIDs) == 0 {
		return []*MemberBadgeRow{}, nil
	}

	query := `
		SELECT sub.user_id, sub.badge_id, sub.badge_name, sub.granted_at FROM (
			SELECT DISTINCT ON (o.user_id) o.user_id, o.badge_id, b.name AS badge_name, o.granted_at
			FROM badge_ownership o
			JOIN badges b ON b.id = o.badge_id
			WHERE o.user_id = ANY($1)
			ORDER BY o.user_id, o.granted_at DESC
		) sub
		ORDER BY sub.granted_at DESC
		LIMIT $2
	`
	var rows []*MemberBadgeRow
	if err := s.db.Select(&rows, query, pq.Array(userIDs), limit); err != nil {
		return nil, errors.Wrap(err, "failed to get recent member badges")
	}
	if rows == nil {
		rows = []*MemberBadgeRow{}
	}
	return rows, nil
}

type BadgePeriodStats struct {
	TotalCount      int `db:"total_count"`
	UniqueUserCount int `db:"unique_user_count"`
}

func (s *SQLStore) GetBadgePeriodStats(startMs, endMs int64) (*BadgePeriodStats, error) {
	query := `
		SELECT COUNT(*) AS total_count, COUNT(DISTINCT user_id) AS unique_user_count
		FROM badge_ownership
		WHERE granted_at >= $1 AND granted_at <= $2
	`
	var stats BadgePeriodStats
	if err := s.db.Get(&stats, query, startMs, endMs); err != nil {
		return nil, errors.Wrap(err, "failed to get badge period stats")
	}
	return &stats, nil
}

func (s *SQLStore) GetBadgePeriodStatsForUsers(userIDs []string, startMs, endMs int64) (*BadgePeriodStats, error) {
	if len(userIDs) == 0 {
		return &BadgePeriodStats{}, nil
	}
	query := `
		SELECT COUNT(*) AS total_count, COUNT(DISTINCT user_id) AS unique_user_count
		FROM badge_ownership
		WHERE user_id = ANY($1) AND granted_at >= $2 AND granted_at <= $3
	`
	var stats BadgePeriodStats
	if err := s.db.Get(&stats, query, pq.Array(userIDs), startMs, endMs); err != nil {
		return nil, errors.Wrap(err, "failed to get badge period stats for users")
	}
	return &stats, nil
}
