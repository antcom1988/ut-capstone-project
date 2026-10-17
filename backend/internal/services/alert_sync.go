package services

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"sikapj-backend/internal/models"
)

func PersistentAlertStatus(status string) string {
	switch status {
	case "SEGERA":
		return "SEGERA"
	case "DIINGATKAN":
		return "MENDEKATI"
	default:
		return "TERTUTUP"
	}
}

func SyncAlertLogs(db *sqlx.DB, defaultThreshold int) (int, error) {
	threshold, err := AlertThreshold(db, defaultThreshold)
	if err != nil {
		return 0, err
	}
	employees := []models.Pegawai{}
	if err := db.Select(&employees, `SELECT * FROM sikapj.pegawai WHERE is_active=TRUE`); err != nil {
		return 0, err
	}
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	count := 0
	for i := range employees {
		for _, a := range BuildPegawaiAlertItems(&employees[i], threshold) {
			status := PersistentAlertStatus(a.Status)
			_, err = tx.Exec(`INSERT INTO sikapj.alert_log(nip,tipe_alert,tanggal_jatuh_tempo,status,notified_at,updated_at) VALUES($1,$2,$3,$4,CASE WHEN $4::varchar='TERTUTUP' THEN NULL ELSE NOW() END,NOW()) ON CONFLICT(nip,tipe_alert,tanggal_jatuh_tempo) DO UPDATE SET status=EXCLUDED.status,notified_at=CASE WHEN sikapj.alert_log.status IS DISTINCT FROM EXCLUDED.status AND EXCLUDED.status<>'TERTUTUP' THEN NOW() ELSE sikapj.alert_log.notified_at END,updated_at=NOW()`, a.NIP, a.TipeAlert, a.TanggalJatuhTempo, status)
			if err != nil {
				return 0, fmt.Errorf("sync alert: %w", err)
			}
			if _, err = tx.Exec(`UPDATE sikapj.alert_log SET status='TERTUTUP',updated_at=NOW() WHERE nip=$1 AND tipe_alert=$2 AND tanggal_jatuh_tempo<>$3 AND status<>'TERTUTUP'`, a.NIP, a.TipeAlert, a.TanggalJatuhTempo); err != nil {
				return 0, fmt.Errorf("close old alert: %w", err)
			}
			count++
		}
	}
	if _, err = tx.Exec(`UPDATE sikapj.alert_log l SET status='TERTUTUP',updated_at=NOW() WHERE status<>'TERTUTUP' AND NOT EXISTS (SELECT 1 FROM sikapj.pegawai p WHERE p.nip=l.nip AND p.is_active AND (l.tipe_alert<>'KP' OR p.status_pegawai='PNS'))`); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}
