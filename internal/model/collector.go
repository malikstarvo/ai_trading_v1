package model

import "time"

type CollectorHealth struct {
	ServiceName   string     `db:"service_name"`
	Status        string     `db:"status"`
	LastSuccessAt *time.Time `db:"last_success_at"`
	LastErrorAt   *time.Time `db:"last_error_at"`
	LastErrorMsg  string     `db:"last_error_msg"`
	UpdatedAt     time.Time  `db:"updated_at"`
}
