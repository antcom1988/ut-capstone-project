package services

import "testing"

func TestPersistentAlertStatus(t *testing.T) {
	tests := map[string]string{"SEGERA": "SEGERA", "DIINGATKAN": "MENDEKATI", "AMAN": "TERTUTUP"}
	for in, want := range tests {
		if got := PersistentAlertStatus(in); got != want {
			t.Fatalf("%s => %s want %s", in, got, want)
		}
	}
}
