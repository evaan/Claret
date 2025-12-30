package util

import (
	"regexp"
	"strings"
)

func jaroSim(str1, str2 string) float64 {
	if len(str1) == 0 && len(str2) == 0 {
		return 1
	}
	if len(str1) == 0 || len(str2) == 0 {
		return 0
	}
	match_distance := len(str1)
	if len(str2) > match_distance {
		match_distance = len(str2)
	}
	match_distance = match_distance/2 - 1
	str1_matches := make([]bool, len(str1))
	str2_matches := make([]bool, len(str2))
	matches := 0.
	transpositions := 0.
	for i := range str1 {
		start := i - match_distance
		if start < 0 {
			start = 0
		}
		end := i + match_distance + 1
		if end > len(str2) {
			end = len(str2)
		}
		for k := start; k < end; k++ {
			if str2_matches[k] {
				continue
			}
			if str1[i] != str2[k] {
				continue
			}
			str1_matches[i] = true
			str2_matches[k] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	k := 0
	for i := range str1 {
		if !str1_matches[i] {
			continue
		}
		for !str2_matches[k] {
			k++
		}
		if str1[i] != str2[k] {
			transpositions++
		}
		k++
	}
	transpositions /= 2
	return (matches/float64(len(str1)) +
		matches/float64(len(str2)) +
		(matches-transpositions)/matches) / 3
}

func normalizeWords(name string) []string {
	name = strings.TrimSpace(name)
	if strings.Contains(name, ",") {
		parts := strings.SplitN(name, ",", 2)
		name = strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
	}
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^\w\s]`)
	name = re.ReplaceAllString(name, "")
	words := strings.Fields(name)
	return words
}

func mixedNameDistance(s, t string) float64 {
	wordsS := normalizeWords(s)
	wordsT := normalizeWords(t)

	maxSim := 0.0
	used := make([]bool, len(wordsT))

	for _, ws := range wordsS {
		best := 0.0
		bestIndex := -1
		for i, wt := range wordsT {
			if used[i] {
				continue
			}
			sim := jaroSim(ws, wt)
			if sim > best {
				best = sim
				bestIndex = i
			}
		}
		if bestIndex >= 0 {
			used[bestIndex] = true
			maxSim += best
		}
	}

	avgSim := maxSim / float64(len(wordsS))
	return 1.0 - avgSim
}

type NameAndDistance struct {
	Name     string
	Distance float64
}

func ClosestName(names []string, name string) NameAndDistance {
	closest := &NameAndDistance{Name: "", Distance: 0}

	for _, prof := range names {
		dist := 1.0 - mixedNameDistance(prof, name)
		if dist > closest.Distance && dist > 0.9 {
			closest.Name = prof
			closest.Distance = dist
		}
	}

	return *closest
}
