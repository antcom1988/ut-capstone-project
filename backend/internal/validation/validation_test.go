package validation

import (
	"bytes"
	"testing"
	"time"
)

func TestValidNIP(t *testing.T) {
	if !ValidNIP("199003202014022005") {
		t.Fatal("NIP 18 digit seharusnya valid")
	}
	for _, nip := range []string{"19900320201402200", "19900320201402200A", " 199003202014022005"} {
		if ValidNIP(nip) {
			t.Fatalf("NIP %q seharusnya ditolak", nip)
		}
	}
}

func TestParseRequiredDate(t *testing.T) {
	got, err := ParseRequiredDate("2026-10-13", "tanggal")
	if err != nil || got.Format(time.DateOnly) != "2026-10-13" {
		t.Fatalf("tanggal valid gagal diparse: %v, %v", got, err)
	}
	if _, err := ParseRequiredDate("", "tanggal"); err == nil {
		t.Fatal("tanggal kosong seharusnya ditolak")
	}
	if _, err := ParseRequiredDate("13-10-2026", "tanggal"); err == nil {
		t.Fatal("format tanggal salah seharusnya ditolak")
	}
}

func TestCanAccessNIP(t *testing.T) {
	if !CanAccessNIP("ADMIN", "", "pegawai-lain") {
		t.Fatal("admin seharusnya boleh mengakses pegawai")
	}
	if !CanAccessNIP("PEGAWAI", "123", "123") {
		t.Fatal("pegawai seharusnya boleh mengakses data sendiri")
	}
	if CanAccessNIP("PEGAWAI", "123", "456") {
		t.Fatal("pegawai tidak boleh mengakses data pegawai lain")
	}
}

func TestDetectAllowedFileType(t *testing.T) {
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 600)...)
	mime, ext, err := DetectAllowedFileType(bytes.NewReader(pdf), ".pdf")
	if err != nil || mime != "application/pdf" || ext != ".pdf" {
		t.Fatalf("PDF valid ditolak: mime=%q ext=%q err=%v", mime, ext, err)
	}

	fakePDF := append([]byte("<script>alert(1)</script>"), bytes.Repeat([]byte("x"), 600)...)
	if _, _, err := DetectAllowedFileType(bytes.NewReader(fakePDF), ".pdf"); err == nil {
		t.Fatal("file HTML berkedok PDF seharusnya ditolak")
	}
}
