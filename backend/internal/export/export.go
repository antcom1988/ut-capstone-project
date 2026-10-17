package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"
)

type EmployeeRow struct {
	NIP          string
	Nama         string
	Status       string
	KGBDue       string
	KPDue        string
	Completeness float64
}

var headers = []string{"NIP", "Nama Lengkap", "Status Pegawai", "Jatuh Tempo KGB", "Jatuh Tempo KP", "Kelengkapan Berkas (%)"}

func records(rows []EmployeeRow) [][]string {
	out := make([][]string, 0, len(rows)+1)
	out = append(out, headers)
	for _, r := range rows {
		out = append(out, []string{r.NIP, r.Nama, r.Status, r.KGBDue, r.KPDue, strconv.FormatFloat(r.Completeness, 'f', 2, 64)})
	}
	return out
}

func CSV(rows []EmployeeRow) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	data := records(rows)
	for _, row := range data {
		for i, cell := range row {
			trimmed := strings.TrimSpace(cell)
			if strings.ContainsAny(cell, "\t\r\n") || (len(trimmed) > 0 && strings.ContainsRune("=+-@", rune(trimmed[0]))) {
				row[i] = "'" + cell
			}
		}
	}
	if err := w.WriteAll(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func XLSX(rows []EmployeeRow) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Rekap Pegawai"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	for i, record := range records(rows) {
		for j, value := range record {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			if err := f.SetCellValue(sheet, cell, value); err != nil {
				return nil, err
			}
		}
	}
	style, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"D9EAF7"}, Pattern: 1}})
	if err != nil {
		return nil, err
	}
	_ = f.SetCellStyle(sheet, "A1", "F1", style)
	_ = f.SetColWidth(sheet, "A", "F", 22)
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func PDF(rows []EmployeeRow) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetTitle("Rekap Data Pegawai SIKAP-J", false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Rekap Data Pegawai SIKAP-J", "", 1, "C", false, 0, "")
	widths := []float64{42, 55, 30, 42, 42, 45}
	pdf.SetFont("Arial", "B", 8)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 8, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 8)
	for _, record := range records(rows)[1:] {
		for i, v := range record {
			pdf.CellFormat(widths[i], 8, v, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
