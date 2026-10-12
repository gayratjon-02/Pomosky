package views

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

func TestMoodIDsMatchTheFrontend(t *testing.T) {
	src, err := os.ReadFile("../../web/src/sky/moods.ts")
	if err != nil {
		t.Fatal(err)
	}
	// Every mood object in moods.ts starts with `id: "<name>"`.
	var front []string
	for _, m := range regexp.MustCompile(`\bid: "([a-z]+)"`).FindAllStringSubmatch(string(src), -1) {
		if m[1] != "live" {
			front = append(front, m[1])
		}
	}
	back := MoodIDs()
	sort.Strings(front)
	sort.Strings(back)
	if len(front) != len(back) {
		t.Fatalf("frontend moods %v, server moods %v", front, back)
	}
	for i := range front {
		if front[i] != back[i] {
			t.Fatalf("frontend moods %v, server moods %v", front, back)
		}
	}
}

func TestValidMood(t *testing.T) {
	for id, want := range map[string]bool{"dawn": true, "storm": true, "live": false, "": false, "<script>": false, "DAWN": false} {
		if got := ValidMood(id); got != want {
			t.Errorf("ValidMood(%q) = %v, want %v", id, got, want)
		}
	}
}
