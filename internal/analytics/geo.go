package analytics

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/netip"
	"os"
	"sort"
	"strings"
)

// CountryDB maps an IP address to a two-letter country code. It is loaded from
// the free DB-IP "IP to Country Lite" CSV (lines "first,last,CC", IPv4 and IPv6,
// CC BY 4.0, see https://db-ip.com). Only the code is kept: the address that
// was looked up is never stored. IPv6 ranges are matched on the first 64 bits.
type CountryDB struct {
	v4 []range4
	v6 []range6
}

type range4 struct {
	first, last uint32
	cc          [2]byte
}

type range6 struct {
	first, last uint64
	cc          [2]byte
}

// LoadCountryDB reads the CSV at path; a name ending in ".gz" is decompressed.
func LoadCountryDB(path string) (*CountryDB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("geoip: %w", err)
		}
		defer gz.Close()
		r = gz
	}
	return ParseCountryDB(r)
}

// ParseCountryDB parses the CSV. Unknown ("ZZ") ranges are dropped and
// neighbouring ranges of the same country are merged to save memory.
func ParseCountryDB(r io.Reader) (*CountryDB, error) {
	db := &CountryDB{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		cc := strings.ToUpper(strings.TrimSpace(parts[2]))
		if len(cc) != 2 || cc == "ZZ" {
			continue
		}
		first, err1 := netip.ParseAddr(strings.TrimSpace(parts[0]))
		last, err2 := netip.ParseAddr(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		code := [2]byte{cc[0], cc[1]}
		if first.Is4() && last.Is4() {
			db.v4 = appendRange4(db.v4, range4{first: v4Num(first), last: v4Num(last), cc: code})
		} else if first.Is6() && last.Is6() {
			db.v6 = appendRange6(db.v6, range6{first: v6High(first), last: v6High(last), cc: code})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("geoip: %w", err)
	}
	if len(db.v4)+len(db.v6) == 0 {
		return nil, fmt.Errorf("geoip: no usable ranges")
	}
	sort.Slice(db.v4, func(i, j int) bool { return db.v4[i].first < db.v4[j].first })
	sort.Slice(db.v6, func(i, j int) bool { return db.v6[i].first < db.v6[j].first })
	return db, nil
}

func appendRange4(rs []range4, r range4) []range4 {
	if n := len(rs); n > 0 && rs[n-1].cc == r.cc && rs[n-1].last+1 == r.first {
		rs[n-1].last = r.last
		return rs
	}
	return append(rs, r)
}

func appendRange6(rs []range6, r range6) []range6 {
	if n := len(rs); n > 0 && rs[n-1].cc == r.cc && rs[n-1].last+1 == r.first {
		rs[n-1].last = r.last
		return rs
	}
	return append(rs, r)
}

func v4Num(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func v6High(a netip.Addr) uint64 {
	b := a.As16()
	var n uint64
	for i := 0; i < 8; i++ {
		n = n<<8 | uint64(b[i])
	}
	return n
}

// Country returns the two-letter country of ip, or "" when it is not in the
// table (private, unknown or not an address).
func (db *CountryDB) Country(ip string) string {
	if db == nil {
		return ""
	}
	addr, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return ""
	}
	addr = addr.Unmap()
	if addr.Is4() {
		n := v4Num(addr)
		i := sort.Search(len(db.v4), func(i int) bool { return db.v4[i].last >= n })
		if i < len(db.v4) && db.v4[i].first <= n {
			return string(db.v4[i].cc[:])
		}
		return ""
	}
	n := v6High(addr)
	i := sort.Search(len(db.v6), func(i int) bool { return db.v6[i].last >= n })
	if i < len(db.v6) && db.v6[i].first <= n {
		return string(db.v6[i].cc[:])
	}
	return ""
}
