package realtime

import (
	"encoding/json"
	"fmt"
	"testing"
)

func decodePeople(t *testing.T, b []byte) PeopleData {
	t.Helper()
	env, err := Decode(b)
	if err != nil || env.Type != TypePeople {
		t.Fatalf("decode: %v %q", err, env.Type)
	}
	var d PeopleData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestPeopleListsAccountsOnceAndNicknames(t *testing.T) {
	rh, _ := ownedRoom("u1")
	a1 := NewClient("s1", nil, 8)
	a1.Account, a1.AccountName, a1.AccountAvatar = "u1", "Ayla", "https://x/a.png"
	a2 := NewClient("s2", nil, 8) // the same person in a second tab
	a2.Account, a2.AccountName = "u1", "Ayla"
	n := NewClient("s3", nil, 8)
	n.Nickname = "Bek"
	anon := NewClient("s4", nil, 8) // nameless viewers are not listed
	for _, c := range []*Client{a1, a2, n, anon} {
		rh.Join(c)
	}

	d := decodePeople(t, rh.People())
	if len(d.People) != 2 || d.More != 0 {
		t.Fatalf("people %+v", d)
	}
	names := map[string]string{}
	for _, p := range d.People {
		names[p.Name] = p.AvatarURL
	}
	if names["Ayla"] != "https://x/a.png" || names["Bek"] != "" {
		t.Fatalf("people %+v", d.People)
	}
}

func TestPeopleIsCappedAndCountsTheRest(t *testing.T) {
	rh, _ := ownedRoom("")
	for i := 0; i < MaxPeople+3; i++ {
		c := NewClient(fmt.Sprintf("s%02d", i), nil, 8)
		c.Nickname = fmt.Sprintf("v%02d", i)
		rh.Join(c)
	}
	d := decodePeople(t, rh.People())
	if len(d.People) != MaxPeople || d.More != 3 {
		t.Fatalf("got %d people, more=%d", len(d.People), d.More)
	}
}

func TestAnnounceOnlySendsWhenTheListChanges(t *testing.T) {
	rh, _ := ownedRoom("")
	watcher := NewClient("w", nil, 8)
	rh.Join(watcher)
	rh.Announce() // still nobody named: nothing to say
	if len(watcher.Outbox()) != 0 {
		t.Fatalf("an unchanged list was sent")
	}
	watcher.Nickname = "Dana"
	rh.Announce()
	rh.Announce() // identical to the last one
	if len(watcher.Outbox()) != 1 {
		t.Fatalf("a changed list was not sent")
	}
}
