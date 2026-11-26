package main

import (
	"slices"
	"testing"
)

func TestStringSlice(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"repeated", []string{"a/", "b/"}, []string{"a/", "b/"}},
		{"comma-separated", []string{"a/,b/"}, []string{"a/", "b/"}},
		{"mixed", []string{"a/,b/", "c/"}, []string{"a/", "b/", "c/"}},
		{"empty", []string{""}, nil},
		{"extra commas", []string{"a/,", ",b/"}, []string{"a/", "b/"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s stringSlice
			for _, a := range c.args {
				if err := s.Set(a); err != nil {
					t.Fatal(err)
				}
			}
			if !slices.Equal(s, c.want) {
				t.Errorf("got %q, want %q", s, c.want)
			}
		})
	}
}
