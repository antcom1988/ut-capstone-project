package audit

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

type Service struct{ DB sqlx.Execer }

func (s Service) Log(userID *int, action string, details any, ip string) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`INSERT INTO sikapj.audit_log (user_id, action, details, ip_address) VALUES ($1,$2,$3,$4)`, userID, action, string(payload), ip)
	return err
}
