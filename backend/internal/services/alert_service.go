package services

import (
	"sikapj-backend/internal/models"
	"time"
)

// CalculateKGBDueDate menghitung jatuh tempo KGB:
// Jatuh tempo KGB = Max(TMT KGB Terakhir, TMT Pangkat Terakhir) + 2 Tahun
func CalculateKGBDueDate(tmtKGB, tmtPangkat time.Time) time.Time {
	tmtDasar := tmtKGB
	if tmtPangkat.After(tmtKGB) {
		tmtDasar = tmtPangkat
	}
	return addCalendar(tmtDasar, 2, 0)
}

// CalculateKPDueDate menghitung jatuh tempo KP:
// Jatuh tempo KP = TMT Pangkat Terakhir + 4 Tahun
func CalculateKPDueDate(tmtPangkat time.Time) time.Time {
	return addCalendar(tmtPangkat, 4, 0)
}

// EvaluateAlertStatus mengevaluasi selisih hari/bulan terhadap ambang batas (N bulan)
func EvaluateAlertStatus(dueDate time.Time, thresholdMonths int) (status string, diffMonths float64, diffDays int) {
	return EvaluateAlertStatusAt(time.Now(), dueDate, thresholdMonths)
}

// EvaluateAlertStatusAt uses calendar boundaries rather than average month lengths.
func EvaluateAlertStatusAt(now, dueDate time.Time, thresholdMonths int) (status string, diffMonths float64, diffDays int) {
	now = dateOnly(now)
	dueDate = dateOnly(dueDate)
	diffDays = int(dueDate.Sub(now).Hours() / 24)
	diffMonths = float64(diffDays) / 30.4375
	urgentAt := addCalendar(dueDate, 0, -1)
	warningAt := addCalendar(dueDate, 0, -thresholdMonths)
	switch {
	case !now.Before(urgentAt):
		status = "SEGERA"
	case !now.Before(warningAt):
		status = "DIINGATKAN"
	default:
		status = "AMAN"
	}
	return status, diffMonths, diffDays
}

func dateOnly(t time.Time) time.Time {
	// Database DATE values are civil dates, not elapsed instants across zones/DST.
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func addCalendar(t time.Time, years, months int) time.Time {
	targetFirst := time.Date(t.Year()+years, t.Month()+time.Month(months), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDay := time.Date(targetFirst.Year(), targetFirst.Month()+1, 0, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()).Day()
	day := t.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(targetFirst.Year(), targetFirst.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

// BuildPegawaiAlertItems membuat daftar AlertItem untuk satu pegawai
func BuildPegawaiAlertItems(p *models.Pegawai, thresholdMonths int) []models.AlertItem {
	var alerts []models.AlertItem

	// 1. Alert KGB
	kgbDue := CalculateKGBDueDate(p.TMTKGBTerakhir, p.TMTPangkatTerakhir)
	statusKGB, monthsKGB, daysKGB := EvaluateAlertStatus(kgbDue, thresholdMonths)
	tmtDasarKGB := p.TMTKGBTerakhir
	if p.TMTPangkatTerakhir.After(p.TMTKGBTerakhir) {
		tmtDasarKGB = p.TMTPangkatTerakhir
	}

	alerts = append(alerts, models.AlertItem{
		NIP:               p.NIP,
		NamaLengkap:       p.NamaLengkap,
		Bidang:            p.Bidang,
		PangkatGolongan:   p.Pangkat + " (" + p.Golongan + ")",
		TipeAlert:         "KGB",
		TMTDasar:          tmtDasarKGB,
		TanggalJatuhTempo: kgbDue,
		SisaBulan:         monthsKGB,
		SisaHari:          daysKGB,
		Status:            statusKGB,
	})

	// 2. Alert KP (Khusus PNS)
	if p.StatusPegawai == "PNS" {
		kpDue := CalculateKPDueDate(p.TMTPangkatTerakhir)
		statusKP, monthsKP, daysKP := EvaluateAlertStatus(kpDue, thresholdMonths)
		alerts = append(alerts, models.AlertItem{
			NIP:               p.NIP,
			NamaLengkap:       p.NamaLengkap,
			Bidang:            p.Bidang,
			PangkatGolongan:   p.Pangkat + " (" + p.Golongan + ")",
			TipeAlert:         "KP",
			TMTDasar:          p.TMTPangkatTerakhir,
			TanggalJatuhTempo: kpDue,
			SisaBulan:         monthsKP,
			SisaHari:          daysKP,
			Status:            statusKP,
		})
	}

	return alerts
}
