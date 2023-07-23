package db

import (
	"testing"
)

func TestDB(t *testing.T) {
	testcases := []struct {
		name    string
		configs []string
	}{
		{"nil configs", nil},
		{"empty configs", []string{}},
		{"one config", []string{"config1"}},
		{"many configs", []string{"config1", "config2", "config3"}},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			err := SetInstalled(tt.configs)
			if err != nil {
				t.Fatal(err)
			}
			configs, err := GetInstalled()
			if err != nil {
				t.Fatal(err)
			}
			if !arrayEquals(tt.configs, configs) {
				t.Errorf("expected: %v, got: %v", tt.configs, configs)
			}
		})
	}
}

func arrayEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, it := range a {
		if it != b[i] {
			return false
		}
	}
	return true
}
