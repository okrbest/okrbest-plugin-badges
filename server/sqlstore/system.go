package sqlstore

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

func (s *SQLStore) getSystemValue(q queryer, key string) (string, error) {
	var value string

	err := s.getBuilder(q, &value,
		sq.Select("SValue").
			From("Badges_System").
			Where(sq.Eq{"SKey": key}),
	)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", errors.Wrapf(err, "failed to query system key %s", key)
	}

	return value, nil
}

func (s *SQLStore) setSystemValue(e queryExecer, key, value string) error {
	result, err := s.execBuilder(e,
		sq.Update("Badges_System").
			Set("SValue", value).
			Where(sq.Eq{"SKey": key}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to update system key %s", key)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to get rows affected for system key %s", key)
	}
	if rowsAffected > 0 {
		return nil
	}

	_, err = s.execBuilder(e,
		sq.Insert("Badges_System").
			Columns("SKey", "SValue").
			Values(key, value),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert system key %s", key)
	}

	return nil
}
