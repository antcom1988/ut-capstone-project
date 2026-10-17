package validation

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
	"time"
)

var nipPattern = regexp.MustCompile(`^\d{18}$`)

func ValidNIP(nip string) bool {
	return nipPattern.MatchString(nip)
}

func ParseRequiredDate(value, field string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("%s wajib diisi", field)
	}
	t, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s harus berformat YYYY-MM-DD", field)
	}
	return t, nil
}

func CanAccessNIP(role, currentNIP, targetNIP string) bool {
	return role == "ADMIN" || (role == "PEGAWAI" && currentNIP != "" && currentNIP == targetNIP)
}

func DetectAllowedFileType(r io.Reader, namedExt string) (string, string, error) {
	br := bufio.NewReader(r)
	header, err := br.Peek(512)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return "", "", fmt.Errorf("tidak dapat membaca file: %w", err)
	}
	detected := http.DetectContentType(header)
	if media, _, parseErr := mime.ParseMediaType(detected); parseErr == nil && media != "" {
		detected = media
	}
	namedExt = filepath.Ext(namedExt)
	if namedExt == "" {
		namedExt = filepath.Clean("." + namedExt)
	}
	allowed := map[string]string{
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
	}
	expected, ok := allowed[namedExt]
	if !ok || detected != expected {
		return "", "", fmt.Errorf("format atau isi file tidak valid")
	}
	return expected, namedExt, nil
}
