package moderation

import "testing"

func TestValidateNickname(t *testing.T) {
	cases := map[string]bool{
		"Ali":                             true,
		"해연":                              true,
		"a":                               false,
		"":                                false,
		"this-nickname-is-way-too-long-x": false,
		"bad\x00name":                     false,
	}
	for in, ok := range cases {
		if err := ValidateNickname(in); (err == nil) != ok {
			t.Errorf("ValidateNickname(%q) error = %v, want ok=%v", in, err, ok)
		}
	}
}
