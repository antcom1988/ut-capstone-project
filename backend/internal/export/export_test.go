package export

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	"testing"
)

func sampleRows() []EmployeeRow {
	return []EmployeeRow{{NIP: "123456789012345678", Nama: "Budi", Status: "PNS", KGBDue: "2027-01-01", KPDue: "2029-01-01", Completeness: 75}}
}

func TestCSVProducesParseableCSV(t *testing.T) {
	b, err := CSV(sampleRows())
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(b)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[1][0] != "123456789012345678" {
		t.Fatalf("unexpected records: %#v", records)
	}
}

func TestXLSXProducesValidWorkbookArchive(t *testing.T) {
	b, err := XLSX(sampleRows())
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range zr.File {
		if f.Name == "xl/workbook.xml" {
			found = true
		}
	}
	if !found {
		t.Fatal("xlsx workbook entry missing")
	}
}

func TestPDFProducesRealPDF(t *testing.T) {
	b, err := PDF(sampleRows())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "%PDF-") {
		t.Fatalf("not a PDF: %q", b[:min(8, len(b))])
	}
	if _, err := io.Copy(io.Discard, bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
}
