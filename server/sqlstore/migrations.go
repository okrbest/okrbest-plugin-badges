package sqlstore

import (
	"github.com/blang/semver"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(sqlx.Ext, *SQLStore) error
}

var migrations = []Migration{
	{
		fromVersion: semver.MustParse("0.0.0"),
		toVersion:   semver.MustParse("0.1.0"),
		migrationFunc: func(e sqlx.Ext, sqlStore *SQLStore) error {
			if _, err := e.Exec(`
				CREATE TABLE IF NOT EXISTS Badges_System (
					SKey   VARCHAR(64) PRIMARY KEY,
					SValue VARCHAR(1024) NULL
				);
			`); err != nil {
				return errors.Wrapf(err, "failed creating table Badges_System")
			}

			if _, err := e.Exec(`
				CREATE TABLE IF NOT EXISTS badge_types (
					id         VARCHAR(26) PRIMARY KEY,
					name       VARCHAR(128) NOT NULL,
					frame      TEXT,
					created_by VARCHAR(26) NOT NULL,
					can_grant  JSONB NOT NULL DEFAULT '{}',
					can_create JSONB NOT NULL DEFAULT '{}'
				);
			`); err != nil {
				return errors.Wrapf(err, "failed creating table badge_types")
			}

			if _, err := e.Exec(`
				CREATE TABLE IF NOT EXISTS badges (
					id          VARCHAR(26) PRIMARY KEY,
					name        VARCHAR(128) NOT NULL,
					description VARCHAR(1024),
					image       TEXT NOT NULL,
					image_type  VARCHAR(20) NOT NULL DEFAULT 'emoji',
					multiple    BOOLEAN NOT NULL DEFAULT false,
					type_id     VARCHAR(26) NOT NULL,
					created_by  VARCHAR(26) NOT NULL
				);
			`); err != nil {
				return errors.Wrapf(err, "failed creating table badges")
			}

			if _, err := e.Exec(`
				CREATE INDEX IF NOT EXISTS idx_badges_type ON badges(type_id);
			`); err != nil {
				return errors.Wrapf(err, "failed creating index idx_badges_type")
			}

			if _, err := e.Exec(`
				CREATE INDEX IF NOT EXISTS idx_badges_created_by ON badges(created_by);
			`); err != nil {
				return errors.Wrapf(err, "failed creating index idx_badges_created_by")
			}

			if _, err := e.Exec(`
				CREATE TABLE IF NOT EXISTS badge_ownership (
					id         VARCHAR(26) PRIMARY KEY,
					user_id    VARCHAR(26) NOT NULL,
					badge_id   VARCHAR(26) NOT NULL,
					granted_by VARCHAR(26) NOT NULL,
					reason     TEXT,
					granted_at BIGINT NOT NULL
				);
			`); err != nil {
				return errors.Wrapf(err, "failed creating table badge_ownership")
			}

			if _, err := e.Exec(`
				CREATE INDEX IF NOT EXISTS idx_badge_ownership_user ON badge_ownership(user_id);
			`); err != nil {
				return errors.Wrapf(err, "failed creating index idx_badge_ownership_user")
			}

			if _, err := e.Exec(`
				CREATE INDEX IF NOT EXISTS idx_badge_ownership_granted_at ON badge_ownership(user_id, granted_at);
			`); err != nil {
				return errors.Wrapf(err, "failed creating index idx_badge_ownership_granted_at")
			}

			if _, err := e.Exec(`
				CREATE TABLE IF NOT EXISTS badge_subscriptions (
					type_id    VARCHAR(26) NOT NULL,
					channel_id VARCHAR(26) NOT NULL,
					PRIMARY KEY (type_id, channel_id)
				);
			`); err != nil {
				return errors.Wrapf(err, "failed creating table badge_subscriptions")
			}

			if _, err := e.Exec(`
				CREATE INDEX IF NOT EXISTS idx_badge_subs_channel ON badge_subscriptions(channel_id);
			`); err != nil {
				return errors.Wrapf(err, "failed creating index idx_badge_subs_channel")
			}

			return nil
		},
	},
}
