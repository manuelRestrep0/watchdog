package model

import "time"

type Target struct {
	ID        int64     `json:"id"         db:"id"`
	URL       string    `json:"url"        db:"url"`
	Interval  int       `json:"interval"   db:"interval"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Check struct {
	ID         int64     `json:"id"          db:"id"`
	TargetID   int64     `json:"target_id"   db:"target_id"`
	StatusCode int       `json:"status_code" db:"status_code"`
	Latency    int64     `json:"latency_ms"  db:"latency_ms"`
	Ok         bool      `json:"ok"          db:"ok"`
	CheckedAt  time.Time `json:"checked_at"  db:"checked_at"`
}
