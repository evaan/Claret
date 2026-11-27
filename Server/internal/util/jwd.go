package util

import (
	"html"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func normalizeWords(name string) []string {
	name = html.UnescapeString(strings.TrimSpace(name))
	if before, after, ok := strings.Cut(name, ","); ok {
		name = after + " " + before
	}
	name = strings.Map(func(r rune) rune {
		switch {
		case unicode.Is(unicode.Mn, r), r == '\'', r == '’', r == '-', r == '‐', r == '‑':
			return -1
		case unicode.IsLetter(r):
			return unicode.ToLower(r)
		default:
			return ' '
		}
	}, norm.NFD.String(name))
	words := strings.Fields(name)
	for len(words) > 0 && (words[0] == "dr" || words[0] == "prof" || words[0] == "professor") {
		words = words[1:]
	}
	return words
}

// nameSimilarity requires the same surname and compatible given names. Middle
// names can be omitted or abbreviated, but conflicting initials are not ignored.
func nameSimilarity(left, right string) float64 {
	a, b := normalizeWords(left), normalizeWords(right)
	if len(a) < 2 || len(b) < 2 || a[len(a)-1] != b[len(b)-1] {
		return 0
	}
	score := 0.97
	switch {
	case a[0] == b[0]:
	case (len(a[0]) == 1 || len(b[0]) == 1) && a[0][0] == b[0][0]:
		score = 0.91
	case len(a[0]) >= 3 && len(b[0]) >= 3 && (strings.HasPrefix(a[0], b[0]) || strings.HasPrefix(b[0], a[0])):
		score = 0.94
	default:
		return 0
	}
	am, bm := a[1:len(a)-1], b[1:len(b)-1]
	for i := 0; i < len(am) && i < len(bm); i++ {
		if am[i] != bm[i] && !(strings.HasPrefix(am[i], bm[i]) || strings.HasPrefix(bm[i], am[i])) {
			return 0
		}
	}
	if strings.Join(a, " ") == strings.Join(b, " ") {
		return 1
	}
	return score
}

type NameAndDistance struct {
	Name     string
	Distance float64
}

// ClosestName leaves ambiguous matches unassigned rather than linking a rating
// to the wrong instructor. Name preserves the caller's original spelling.
func ClosestName(names []string, name string) NameAndDistance {
	best := NameAndDistance{}
	ambiguous := false
	for _, candidate := range names {
		score := nameSimilarity(candidate, name)
		if score == 0 {
			continue
		}
		if score > best.Distance {
			best = NameAndDistance{Name: candidate, Distance: score}
			ambiguous = false
		} else if score == best.Distance {
			ambiguous = true
		}
	}
	if ambiguous {
		best.Name = ""
	}
	return best
}
