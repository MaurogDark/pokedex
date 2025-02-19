package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Capital Captain Crunch",
			expected: []string{"capital", "captain", "crunch"},
		},
		{
			input:    " Charmander Bulbasaur PIKACHU ",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("Length doesn't match, expected %d but got %d", len(c.expected), len(actual))
		} else {
			for i := range actual {
				word := actual[i]
				expectedWord := c.expected[i]
				if word != expectedWord {
					t.Errorf("Words doen't match, expected %s but got %s", expectedWord, word)
					break
				}
			}
		}
	}
}
