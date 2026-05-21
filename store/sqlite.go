package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/manuelRestrep0/watchdog/model"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sqlx.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func migrate(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS targets (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			url        TEXT NOT NULL,
			interval   INTEGER NOT NULL DEFAULT 30,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS checks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			target_id   INTEGER NOT NULL,
			status_code INTEGER,
			latency_ms  INTEGER,
			ok          BOOLEAN,
			checked_at  DATETIME NOT NULL,
			FOREIGN KEY (target_id) REFERENCES targets(id)
		);
	`)
	return err
}

func (s *SQLiteStore) CreateTarget(url string, interval int) (*model.Target, error) {
	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO targets (url, interval, created_at) VALUES (?, ?, ?)`,
		url, interval, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	return &model.Target{ID: id, URL: url, Interval: interval, CreatedAt: now}, nil
}

func (s *SQLiteStore) ListTargets() ([]model.Target, error) {
	var targets []model.Target
	err := s.db.Select(&targets, `SELECT * FROM targets`)
	return targets, err
}

func (s *SQLiteStore) DeleteTarget(id int64) error {
	_, err := s.db.Exec(`DELETE FROM targets WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) SaveCheck(c *model.Check) error {
	_, err := s.db.Exec(
		`INSERT INTO checks (target_id, status_code, latency_ms, ok, checked_at) VALUES (?, ?, ?, ?, ?)`,
		c.TargetID, c.StatusCode, c.Latency, c.Ok, c.CheckedAt,
	)
	return err
}

func (s *SQLiteStore) GetHistory(targetID int64) ([]model.Check, error) {
	var checks []model.Check
	err := s.db.Select(&checks,
		`SELECT * FROM checks WHERE target_id = ? ORDER BY checked_at DESC LIMIT 100`,
		targetID,
	)
	return checks, err
}
