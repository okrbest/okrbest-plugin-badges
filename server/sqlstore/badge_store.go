package sqlstore

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type BadgeRow struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Image       string `db:"image"`
	ImageType   string `db:"image_type"`
	Multiple    bool   `db:"multiple"`
	TypeID      string `db:"type_id"`
	CreatedBy   string `db:"created_by"`
}

func (r *BadgeRow) ToBadge() *badgesmodel.Badge {
	return &badgesmodel.Badge{
		ID:          badgesmodel.BadgeID(r.ID),
		Name:        r.Name,
		Description: r.Description,
		Image:       r.Image,
		ImageType:   badgesmodel.ImageType(r.ImageType),
		Multiple:    r.Multiple,
		Type:        badgesmodel.BadgeType(r.TypeID),
		CreatedBy:   r.CreatedBy,
	}
}

var badgeColumns = []string{
	"id", "name", "description", "image", "image_type", "multiple", "type_id", "created_by",
}

func (s *SQLStore) InsertBadge(b *badgesmodel.Badge) error {
	_, err := s.execBuilder(s.db,
		sq.Insert("badges").
			Columns(badgeColumns...).
			Values(
				string(b.ID), b.Name, b.Description, b.Image,
				string(b.ImageType), b.Multiple, string(b.Type), b.CreatedBy,
			),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert badge %s", b.ID)
	}
	return nil
}

func (s *SQLStore) GetBadge(id badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	var row BadgeRow
	err := s.getBuilder(s.db, &row,
		sq.Select(badgeColumns...).
			From("badges").
			Where(sq.Eq{"id": string(id)}),
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("badge not found")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get badge %s", id)
	}
	return row.ToBadge(), nil
}

func (s *SQLStore) GetAllBadges() ([]*badgesmodel.Badge, error) {
	var rows []*BadgeRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select(badgeColumns...).
			From("badges"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all badges")
	}

	badges := make([]*badgesmodel.Badge, 0, len(rows))
	for _, r := range rows {
		badges = append(badges, r.ToBadge())
	}
	return badges, nil
}

func (s *SQLStore) UpdateBadge(b *badgesmodel.Badge) error {
	result, err := s.execBuilder(s.db,
		sq.Update("badges").
			Set("name", b.Name).
			Set("description", b.Description).
			Set("image", b.Image).
			Set("image_type", string(b.ImageType)).
			Set("multiple", b.Multiple).
			Set("type_id", string(b.Type)).
			Where(sq.Eq{"id": string(b.ID)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update badge %s", b.ID)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for badge update")
	}
	if rowsAffected == 0 {
		return errors.New("badge not found")
	}
	return nil
}

func (s *SQLStore) DeleteBadge(id badgesmodel.BadgeID) error {
	_, err := s.execBuilder(s.db,
		sq.Delete("badges").
			Where(sq.Eq{"id": string(id)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete badge %s", id)
	}
	return nil
}

func (s *SQLStore) GetBadgesByType(typeID badgesmodel.BadgeType) ([]*badgesmodel.Badge, error) {
	var rows []*BadgeRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select(badgeColumns...).
			From("badges").
			Where(sq.Eq{"type_id": string(typeID)}),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get badges by type %s", typeID)
	}

	badges := make([]*badgesmodel.Badge, 0, len(rows))
	for _, r := range rows {
		badges = append(badges, r.ToBadge())
	}
	return badges, nil
}

func (s *SQLStore) GetBadgesByCreator(botID string) ([]*badgesmodel.Badge, error) {
	var rows []*BadgeRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select(badgeColumns...).
			From("badges").
			Where(sq.Eq{"created_by": botID}),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get badges by creator %s", botID)
	}

	badges := make([]*badgesmodel.Badge, 0, len(rows))
	for _, r := range rows {
		badges = append(badges, r.ToBadge())
	}
	return badges, nil
}

func (s *SQLStore) NewBadgeID() string {
	return model.NewId()
}
