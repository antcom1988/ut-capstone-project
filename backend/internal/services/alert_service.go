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
	return tmtDasar.AddDate(2, 0, 0)
}

// CalculateKPDueDate menghitung jatuh tempo KP:
// Jatuh tempo KP = TMT Pangkat Terakhir + 4 Tahun
func CalculateKPDueDate(tmtPangkat time.Time) time.Time {
	return tmtPangkat.AddDate(4, 0, 0)
}

// EvaluateAlertStatus mengevaluasi selisih hari/bulan terhadap ambang batas (N bulan)
func EvaluateAlertStatus(dueDate time.Time, thresholdMonths int) (status string, diffMonths float64, diffDays int) {
	now := time.Now()
	diffDuration := dueDate.Sub(now)
	diffDays = int(diffDuration.Hours() / 24)
	diffMonths = float64(diffDays) / 30.4375 // rata-rata hari per bulan

	if diffDays <= 30 {
		status = "SEGERA" // 🔴 Sangat mendesak (kurang dari / sama dengan 1 bulan)
	} else if diffMonths <= float64(thresholdMonths) {
		status = "DIINGATKAN" // 🟡 Masuk rentang N bulan pengingat
	} else {
		status = "AMAN" // 🟢 Masih aman
	}

	return status, diffMonths, diffDays
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
