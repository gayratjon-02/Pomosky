package analytics

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleCSV = `0.0.0.0,0.255.255.255,ZZ
1.0.0.0,1.0.0.255,AU
1.0.1.0,1.0.3.255,CN
1.0.4.0,1.0.7.255,CN
20.24.64.0,20.24.127.255,HK
::,1fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff,ZZ
2001:db8::,2001:db8:0:ffff:ffff:ffff:ffff:ffff,KR
`

func TestCountryLookup(t *testing.T) {
	db, err := ParseCountryDB(strings.NewReader(sampleCSV))
	if err != nil {
		t.Fatal(err)
	}
	for ip, want := range map[string]string{
		"1.0.0.7":              "AU",
		"1.0.2.9":              "CN",
		"1.0.5.1":              "CN", // a merged neighbour range
		"1.0.8.1":              "",   // past the last range
		"20.24.100.1":          "HK",
		"0.1.2.3":              "", // ZZ is unknown
		"::ffff:20.24.100.1":   "HK",
		"2001:db8:0:1:2:3:4:5": "KR",
		"2001:db8:1::1":        "",
		"10.0.0.1":             "",
		"not an address":       "",
		"":                     "",
	} {
		if got := db.Country(ip); got != want {
			t.Errorf("Country(%q) = %q, want %q", ip, got, want)
		}
	}
	if len(db.v4) != 3 { // AU, one merged CN range, HK
		t.Errorf("ranges were not merged: %d", len(db.v4))
	}
}

func TestCountryDBFromGzipFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "geo.csv.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	_, _ = gz.Write([]byte(sampleCSV))
	_ = gz.Close()
	_ = f.Close()
	db, err := LoadCountryDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := db.Country("1.0.0.1"); got != "AU" {
		t.Fatalf("got %q", got)
	}
}

func TestCountryDBRejectsEmptyAndNilIsSafe(t *testing.T) {
	if _, err := ParseCountryDB(strings.NewReader("garbage\n")); err == nil {
		t.Fatal("an empty table must be an error")
	}
	var db *CountryDB
	if db.Country("1.0.0.1") != "" {
		t.Fatal("a nil table must answer empty")
	}
}
