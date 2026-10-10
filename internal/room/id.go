package room

// ID identifies a room. Generated IDs are 8 characters of lowercase
// base32 (a-z, 2-7); see platform/idgen.
type ID string

// String returns the ID as a plain string.
func (id ID) String() string { return string(id) }

// Valid reports whether id has the shape of a generated room ID.
func (id ID) Valid() bool {
	if len(id) != 8 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z') && !(c >= '2' && c <= '7') {
			return false
		}
	}
	return true
}
