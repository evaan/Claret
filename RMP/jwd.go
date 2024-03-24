package main

import (
	"math"
)

//https://rosettacode.org/wiki/Jaro-Winkler_distance#Go

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

// modified jaro winkler distance to count the first and last part of the name, prioritising the end.
func mixedJaroWinklerDist(s, t string) float64 {
	ls := len(s)
	lt := len(t)
	lmax := lt
	if ls < lt {
		lmax = ls
	}
	l := 0.0
	for i := 1; i < int(math.Min(float64(6), float64(lmax))); i++ {
		if s[len(s)-i] == t[len(t)-i] {
			l++
		}
	}
	for i := 1; i < int(math.Min(float64(2), float64(lmax))); i++ {
		if s[i] != t[i] {
			l -= 2
		}
	}
	js := jaroSim(s, t)
	p := 0.1
	ws := (1 - js) - float64(l)*p*(1-js)
	return ws
}

type NameAndDistance struct {
	Name     string
	Distance float64
}

func closestName(names []string, name string) NameAndDistance {
	closest := &NameAndDistance{Name: "", Distance: 2}

	for _, prof := range names {
		dist := mixedJaroWinklerDist(prof, name)
		if dist < closest.Distance && dist < 0.1 {
			closest.Name = prof
			closest.Distance = dist
		}
	}

	return *closest
}
