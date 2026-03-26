package internal

import "strings"

func ToTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	s = strings.ToLower(s)
	words := strings.Split(s, " ")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}

	}
	return strings.Join(words, " ")
}
