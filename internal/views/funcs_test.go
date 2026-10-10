package views

import "testing"

func TestInitial(t *testing.T) {
	cases := map[string]string{"ayla": "A", "  Tik Tok": "T", "élan": "É", "": "?", "   ": "?", "7zip": "7"}
	for in, want := range cases {
		if got := initial(in); got != want {
			t.Errorf("initial(%q) = %q, want %q", in, got, want)
		}
	}
}
