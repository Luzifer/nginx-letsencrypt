package main

import "testing"

func Test_Regex(t *testing.T) {
	var (
		lines = map[string]bool{
			"  #   server_name 1.example.com;": false,
			"      server_name 2.example.com;": true,
			"server_name 3.example.com;":       true,
		}
	)

	for line, result := range lines {
		if realResult := serverLine.MatchString(line); realResult != result {
			t.Errorf("Line %q had result %v (expected %v)", line, realResult, result)
		}
	}
}
