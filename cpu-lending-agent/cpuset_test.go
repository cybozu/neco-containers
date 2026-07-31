package main

import "testing"

func TestParseCPUSetAndString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		input   string
		want    string // canonical round-trip representation
		wantErr bool
	}{
		{name: "empty", input: "", want: ""},
		{name: "whitespace", input: " \n", want: ""},
		{name: "single", input: "3", want: "3"},
		{name: "range", input: "0-3", want: "0-3"},
		{name: "mixed", input: "0-1,4-11", want: "0-1,4-11"},
		{name: "non canonical order", input: "4-11,0-1", want: "0-1,4-11"},
		{name: "overlapping append", input: "0-3,2", want: "0-3"},
		{name: "adjacent merge", input: "0-1,2,3-4", want: "0-4"},
		{name: "garbage", input: "abc", wantErr: true},
		{name: "descending range", input: "5-2", wantErr: true},
		{name: "trailing comma", input: "1,", wantErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			set, err := ParseCPUSet(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseCPUSet(%q) succeeded, want error", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCPUSet(%q): %v", tc.input, err)
			}
			if got := set.String(); got != tc.want {
				t.Errorf("ParseCPUSet(%q).String() = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func mustParse(t *testing.T, s string) CPUSet {
	t.Helper()
	set, err := ParseCPUSet(s)
	if err != nil {
		t.Fatalf("ParseCPUSet(%q): %v", s, err)
	}
	return set
}

func TestCPUSetUnion(t *testing.T) {
	t.Parallel()
	got := mustParse(t, "0-1,6-11").Union(mustParse(t, "2-3"))
	if want := "0-3,6-11"; got.String() != want {
		t.Errorf("Union = %q, want %q", got.String(), want)
	}
}

func TestCPUSetEqual(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		a, b string
		want bool
	}{
		{name: "equal different notation", a: "0-2", b: "0,1,2", want: true},
		{name: "different members", a: "0-2", b: "0-3", want: false},
		{name: "both empty", a: "", b: "", want: true},
		{name: "one empty", a: "0", b: "", want: false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := mustParse(t, tc.a).Equal(mustParse(t, tc.b)); got != tc.want {
				t.Errorf("Equal(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
