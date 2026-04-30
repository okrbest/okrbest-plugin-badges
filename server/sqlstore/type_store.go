package sqlstore

import (
	"database/sql"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type BadgeTypeRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Frame     string `db:"frame"`
	CreatedBy string `db:"created_by"`
	CanGrant  []byte `db:"can_grant"`
	CanCreate []byte `db:"can_create"`
}

func (r *BadgeTypeRow) ToBadgeTypeDefinition() (*badgesmodel.BadgeTypeDefinition, error) {
	t := &badgesmodel.BadgeTypeDefinition{
		ID:        badgesmodel.BadgeType(r.ID),
		Name:      r.Name,
		Frame:     r.Frame,
		CreatedBy: r.CreatedBy,
	}

	if len(r.CanGrant) > 0 {
		if err := json.Unmarshal(r.CanGrant, &t.CanGrant); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal can_grant")
		}
	}
	if len(r.CanCreate) > 0 {
		if err := json.Unmarshal(r.CanCreate, &t.CanCreate); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal can_create")
		}
	}

	return t, nil
}

func marshalPermissionScheme(ps badgesmodel.PermissionScheme) ([]byte, error) {
	b, err := json.Marshal(ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal permission scheme")
	}
	return b, nil
}

func (s *SQLStore) InsertType(t *badgesmodel.BadgeTypeDefinition) error {
	canGrant, err := marshalPermissionScheme(t.CanGrant)
	if err != nil {
		return err
	}
	canCreate, err := marshalPermissionScheme(t.CanCreate)
	if err != nil {
		return err
	}

	_, err = s.execBuilder(s.db,
		sq.Insert("badge_types").
			Columns("id", "name", "frame", "created_by", "can_grant", "can_create").
			Values(string(t.ID), t.Name, t.Frame, t.CreatedBy, canGrant, canCreate),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert type %s", t.ID)
	}
	return nil
}

func (s *SQLStore) GetType(id badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error) {
	var row BadgeTypeRow
	err := s.getBuilder(s.db, &row,
		sq.Select("id", "name", "frame", "created_by", "can_grant", "can_create").
			From("badge_types").
			Where(sq.Eq{"id": string(id)}),
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("type not found")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get type %s", id)
	}
	return row.ToBadgeTypeDefinition()
}

func (s *SQLStore) GetAllTypes() (badgesmodel.BadgeTypeList, error) {
	var rows []*BadgeTypeRow
	err := s.selectBuilder(s.db, &rows,
		sq.Select("id", "name", "frame", "created_by", "can_grant", "can_create").
			From("badge_types"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all types")
	}

	types := make(badgesmodel.BadgeTypeList, 0, len(rows))
	for _, r := range rows {
		t, err := r.ToBadgeTypeDefinition()
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, nil
}

func (s *SQLStore) UpdateType(t *badgesmodel.BadgeTypeDefinition) error {
	canGrant, err := marshalPermissionScheme(t.CanGrant)
	if err != nil {
		return err
	}
	canCreate, err := marshalPermissionScheme(t.CanCreate)
	if err != nil {
		return err
	}

	result, err := s.execBuilder(s.db,
		sq.Update("badge_types").
			Set("name", t.Name).
			Set("frame", t.Frame).
			Set("can_grant", canGrant).
			Set("can_create", canCreate).
			Where(sq.Eq{"id": string(t.ID)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update type %s", t.ID)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for type update")
	}
	if rowsAffected == 0 {
		return errors.New("type not found")
	}
	return nil
}

func (s *SQLStore) DeleteType(id badgesmodel.BadgeType) error {
	_, err := s.execBuilder(s.db,
		sq.Delete("badge_types").
			Where(sq.Eq{"id": string(id)}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to delete type %s", id)
	}
	return nil
}

func (s *SQLStore) GetTypeByCreator(botID string) (*badgesmodel.BadgeTypeDefinition, error) {
	var row BadgeTypeRow
	err := s.getBuilder(s.db, &row,
		sq.Select("id", "name", "frame", "created_by", "can_grant", "can_create").
			From("badge_types").
			Where(sq.Eq{"created_by": botID}).
			Limit(1),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get type by creator %s", botID)
	}
	return row.ToBadgeTypeDefinition()
}

func (s *SQLStore) NewTypeID() string {
	return model.NewId()
}
