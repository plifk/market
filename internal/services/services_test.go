package services

import (
	"testing"
)

func TestNew11RandomID(t *testing.T) {
	const (
		alphabet = "123456789ABCDEFGHJKLMNPQRSTUWVXYZabcdefghijkmnopqrstuwvxyz" // base58
		size     = 11
	)

	var alpha = map[rune]struct{}{}
	var duplicates = map[string]struct{}{}
	for _, char := range alphabet {
		alpha[char] = struct{}{}
	}
	for x := 0; x < 400; x++ {
		got := new11RandomID()
		for _, char := range got {
			if _, ok := alpha[char]; !ok {
				t.Errorf("got id %q with chars not in the alphabet %q", got, alphabet)
			}
		}
		if len(got) != size {
			t.Errorf("got id %q, wanted id with with length %d", got, size)
		}
		if _, ok := duplicates[got]; ok {
			t.Errorf("id generated more than once")
		}
		duplicates[got] = struct{}{}
	}
}
