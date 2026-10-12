package i18n

import "testing"

func TestPick(t *testing.T) {
	cases := []struct {
		cookie, accept string
		want           Lang
	}{
		{"ru", "ko-KR,ko;q=0.9", RU},                   // the cookie wins
		{"xx", "uz-UZ,uz;q=0.9,ru;q=0.8,en;q=0.7", UZ}, // unknown cookie: the header
		{"", "ko-KR,ko;q=0.9,en-US;q=0.8", KO},
		{"", "en-US,en;q=0.9,uz;q=0.8", EN},
		{"", "ru;q=0.4, uz;q=0.9", UZ}, // quality values decide
		{"", "fr-FR,de;q=0.8", EN},     // nothing supported
		{"", "", EN},
		{"", "uz;q=0", EN}, // q=0 means "not acceptable"
	}
	for _, c := range cases {
		if got := Pick(c.cookie, c.accept); got != c.want {
			t.Errorf("Pick(%q, %q) = %s, want %s", c.cookie, c.accept, got, c.want)
		}
	}
}

func TestTranslateFallsBackToEnglish(t *testing.T) {
	dict[UZ] = map[string]string{"Hello": "Salom"}
	defer func() { dict[UZ] = map[string]string{} }()
	if T(UZ, "Hello") != "Salom" || T(UZ, "Missing") != "Missing" || T(EN, "Hello") != "Hello" || T(KO, "Hello") != "Hello" {
		t.Fatal("unexpected translation")
	}
}

func TestParse(t *testing.T) {
	if Parse("ko") != KO || Parse("xx") != EN || Parse("") != EN {
		t.Fatal("Parse")
	}
}
