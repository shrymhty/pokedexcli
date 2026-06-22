package main

import (
	"testing"
	"reflect"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input 	string
		expected []string
	} {
		{
			input: "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			input: "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input: "    HellO WorlD    ",
			expected: []string{"hello", "world"},
		},
	}

	for _, tc := range cases {
		actual := cleanInput(tc.input)

		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("got %v, want %v", actual, tc.expected)
		}
	}
}