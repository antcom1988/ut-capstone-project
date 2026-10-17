package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sikapj-backend/internal/export"
	"sikapj-backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type AdminHandler struct{ DB *sqlx.DB }

type auditEntry struct {
	ID        int64     `db:"id" json:"id"`
	UserID    *int      `db:"user_id" json:"user_id"`
	Username  *string   `db:"username" json:"username"`
	Action    string    `db:"action" json:"action"`
	Details   *string   `db:"details" json:"details"`
	IPAddress *string   `db:"ip_address" json:"ip_address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func parseAuditBoundary(value string, endOfDate bool) (time.Time, bool, error) {
	if parsed, err := time.ParseInLocation(time.DateOnly, value, time.Local); err == nil {
		if endOfDate {
			parsed = parsed.AddDate(0, 0, 1)
		}
		return parsed, false, nil
	}
	for _, layout := range []string{"2006-01-02T15:04", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("format tanggal/waktu tidak valid")
}

func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	action := strings.TrimSpace(c.Query("action"))
	user := strings.TrimSpace(c.Query("user"))
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))
	args := []any{}
	where := []string{"1=1"}
	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}
	if action != "" {
		add("a.action ILIKE '%%' || $%d || '%%'", action)
	}
	if user != "" {
		add("u.username ILIKE '%%' || $%d || '%%'", user)
	}
	if from != "" {
		parsed, _, err := parseAuditBoundary(from, false)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Tanggal/jam mulai tidak valid"})
		}
		add("a.created_at >= $%d", parsed)
	}
	if to != "" {
		parsed, hasTime, err := parseAuditBoundary(to, true)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Tanggal/jam selesai tidak valid"})
		}
		operator := "<"
		if hasTime {
			operator = "<="
		}
		add("a.created_at "+operator+" $%d", parsed)
	}
	var total int
	countQ := `SELECT COUNT(*) FROM sikapj.audit_log a LEFT JOIN sikapj.users u ON u.id=a.user_id WHERE ` + strings.Join(where, " AND ")
	if err := h.DB.Get(&total, countQ, args...); err != nil {
		return internalError(c)
	}
	listArgs := append(append([]any{}, args...), limit, (page-1)*limit)
	q := `SELECT a.id,a.user_id,u.username,a.action,a.details,a.ip_address,a.created_at FROM sikapj.audit_log a LEFT JOIN sikapj.users u ON u.id=a.user_id WHERE ` + strings.Join(where, " AND ") + fmt.Sprintf(" ORDER BY a.created_at DESC,a.id DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	entries := []auditEntry{}
	if err := h.DB.Select(&entries, q, listArgs...); err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"success": true, "data": entries, "pagination": fiber.Map{"page": page, "limit": limit, "total": total, "pages": (total + limit - 1) / limit}})
}

func (h *AdminHandler) GetPendingVerification(c *fiber.Ctx) error {
	type item struct {
		ID         int       `db:"id" json:"id"`
		NIP        string    `db:"nip" json:"nip"`
		Nama       string    `db:"nama_lengkap" json:"nama_lengkap"`
		Dokumen    string    `db:"nama_dokumen" json:"nama_dokumen"`
		FileName   string    `db:"file_name" json:"file_name"`
		UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
	}
	out := []item{}
	err := h.DB.Select(&out, `SELECT b.id,b.nip,p.nama_lengkap,m.nama_dokumen,b.file_name,b.uploaded_at FROM sikapj.e_berkas b JOIN sikapj.pegawai p ON p.nip=b.nip JOIN sikapj.master_dokumen m ON m.id=b.dokumen_id WHERE b.status_verifikasi='Menunggu Verifikasi' ORDER BY b.uploaded_at ASC`)
	if err != nil {
		return internalError(c)
	}
	return c.JSON(fiber.Map{"success": true, "data": out, "total": len(out)})
}

func (h *AdminHandler) ExportEmployees(c *fiber.Ctx) error {
	type dbRow struct {
		NIP        string    `db:"nip"`
		Nama       string    `db:"nama_lengkap"`
		Status     string    `db:"status_pegawai"`
		TMTKGB     time.Time `db:"tmt_kgb_terakhir"`
		TMTPangkat time.Time `db:"tmt_pangkat_terakhir"`
		Approved   int       `db:"approved"`
		Required   int       `db:"required"`
	}
	rows := []dbRow{}
	q := `SELECT p.nip,p.nama_lengkap,p.status_pegawai,p.tmt_kgb_terakhir,p.tmt_pangkat_terakhir,COUNT(DISTINCT b.dokumen_id) FILTER (WHERE b.status_verifikasi='Disetujui' AND ` + services.RequiredDocumentPredicateSQL + `) approved,(SELECT COUNT(*) FROM sikapj.master_dokumen m WHERE ` + services.RequiredDocumentPredicateSQL + `) required FROM sikapj.pegawai p LEFT JOIN sikapj.e_berkas b ON b.nip=p.nip LEFT JOIN sikapj.master_dokumen m ON m.id=b.dokumen_id WHERE p.is_active GROUP BY p.nip ORDER BY p.nama_lengkap`
	if err := h.DB.Select(&rows, q); err != nil && err != sql.ErrNoRows {
		return internalError(c)
	}
	data := make([]export.EmployeeRow, 0, len(rows))
	for _, r := range rows {
		comp := 0.0
		if r.Required > 0 {
			comp = float64(r.Approved) * 100 / float64(r.Required)
		}
		kp := "-"
		if r.Status == "PNS" {
			kp = services.CalculateKPDueDate(r.TMTPangkat).Format(time.DateOnly)
		}
		data = append(data, export.EmployeeRow{NIP: r.NIP, Nama: r.Nama, Status: r.Status, KGBDue: services.CalculateKGBDueDate(r.TMTKGB, r.TMTPangkat).Format(time.DateOnly), KPDue: kp, Completeness: comp})
	}
	format := strings.ToLower(c.Query("format", "xlsx"))
	var body []byte
	var err error
	var contentType, ext string
	switch format {
	case "csv":
		body, err = export.CSV(data)
		contentType = "text/csv; charset=utf-8"
		ext = "csv"
	case "xlsx":
		body, err = export.XLSX(data)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		ext = "xlsx"
	case "pdf":
		body, err = export.PDF(data)
		contentType = "application/pdf"
		ext = "pdf"
	default:
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Format harus csv, xlsx, atau pdf"})
	}
	if err != nil {
		return internalError(c)
	}
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="rekap-pegawai.%s"`, ext))
	return c.Send(body)
}

func internalError(c *fiber.Ctx) error {
	return c.Status(500).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan internal"})
}
