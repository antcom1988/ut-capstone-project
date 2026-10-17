package export

import (
	"bytes"
	"encoding/csv"
	"testing"
)

func TestCSVNeutralizesFormulaCells(t *testing.T) {
	for _, s := range []string{"=1+1", "+cmd", "-1+2", "@SUM(A1)", "\t=1", " \r=1"} {
		b, e := CSV([]EmployeeRow{{Nama: s}})
		if e != nil {
			t.Fatal(e)
		}
		r, e := csv.NewReader(bytes.NewReader(b)).ReadAll()
		if e != nil {
			t.Fatal(e)
		}
		if r[1][1] != "'"+s {
			t.Errorf("unsafe cell %q", r[1][1])
		}
	}
}
