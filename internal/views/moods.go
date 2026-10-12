package views

// MoodCookie is the cookie that remembers the visitor's mood (see web/src/sky/moods.ts).
const MoodCookie = "pomosky_mood"

// moodIDs are the moods the frontend offers, besides the live sky ("live" is the default
// and is not written into the page). A test keeps this list equal to web/src/sky/moods.ts.
var moodIDs = map[string]bool{
	"dawn": true, "clear": true, "golden": true, "sunset": true, "blue": true,
	"night": true, "moon": true, "rain": true, "snow": true, "storm": true,
}

// ValidMood reports whether id is a mood the page can be rendered with.
func ValidMood(id string) bool { return moodIDs[id] }

// MoodIDs returns the ids of all moods except "live" (for tests).
func MoodIDs() []string {
	out := make([]string, 0, len(moodIDs))
	for id := range moodIDs {
		out = append(out, id)
	}
	return out
}
