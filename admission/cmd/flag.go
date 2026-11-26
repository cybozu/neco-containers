package main

import "strings"

// stringSlice is a flag.Value that accepts the flag multiple times and comma-separated values.
type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

// Set ignores empty elements; an empty prefix would match every image.
func (s *stringSlice) Set(v string) error {
	for e := range strings.SplitSeq(v, ",") {
		if e != "" {
			*s = append(*s, e)
		}
	}
	return nil
}
