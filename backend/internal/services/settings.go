package services

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// AlertThreshold falls back only when the setting is absent, never on DB errors.
func AlertThreshold(db *sqlx.DB, fallback int) (int, error) {
	var raw string
	err := db.Get(&raw, `SELECT value FROM sikapj.settings WHERE key = 'ALERT_N_MONTHS'`)
	if err == sql.ErrNoRows {
		if fallback <= 0 {
			return 0, fmt.Errorf("invalid default alert threshold")
		}
		return fallback, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read alert threshold: %w", err)
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid alert threshold %q", raw)
	}
	return n, nil
}
