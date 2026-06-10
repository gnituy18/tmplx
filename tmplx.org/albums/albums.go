// Package albums is the landing-page demo "database": an in-memory table
// queried by tmplx components the same way a real database would be.
package albums

import "strings"

type Album struct {
	Title  string
	Artist string
	Year   int
}

var table = []Album{
	{"Kind of Blue", "Miles Davis", 1959},
	{"A Love Supreme", "John Coltrane", 1965},
	{"Abbey Road", "The Beatles", 1969},
	{"The Dark Side of the Moon", "Pink Floyd", 1973},
	{"Rumours", "Fleetwood Mac", 1977},
	{"Thriller", "Michael Jackson", 1982},
	{"Purple Rain", "Prince", 1984},
	{"Nevermind", "Nirvana", 1991},
	{"OK Computer", "Radiohead", 1997},
	{"Discovery", "Daft Punk", 2001},
}

func Search(q string) []Album {
	q = strings.ToLower(strings.TrimSpace(q))
	var out []Album
	for _, a := range table {
		if strings.Contains(strings.ToLower(a.Title+" "+a.Artist), q) {
			out = append(out, a)
		}
	}
	return out
}
