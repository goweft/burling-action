package main

import (
	"reflect"
	"testing"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		token   string
		strict  bool
		want    []string
	}{
		{"lint default", "lint", "tok.jwt", false, []string{"lint", "--format", "sarif", "tok.jwt"}},
		{"lint strict", "lint", "tok.jwt", true, []string{"lint", "--format", "sarif", "--strict", "tok.jwt"}},
		{"validate", "validate", "a.jwt", false, []string{"validate", "--format", "sarif", "a.jwt"}},
		{"validate-identity", "validate-identity", "id.json", false, []string{"validate-identity", "--format", "sarif", "id.json"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildArgs(tc.command, tc.token, tc.strict); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("buildArgs(%q, %q, %v) = %v, want %v", tc.command, tc.token, tc.strict, got, tc.want)
			}
		})
	}
}

func TestShouldFail(t *testing.T) {
	tests := []struct {
		name        string
		failOnError bool
		code        int
		want        bool
	}{
		{"disabled clean", false, 0, false},
		{"disabled findings", false, 1, false},
		{"enabled clean", true, 0, false},
		{"enabled findings", true, 1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldFail(tc.failOnError, tc.code); got != tc.want {
				t.Errorf("shouldFail(%v, %d) = %v, want %v", tc.failOnError, tc.code, got, tc.want)
			}
		})
	}
}

func TestIsTrue(t *testing.T) {
	for _, s := range []string{"true", "1", "yes"} {
		if !isTrue(s) {
			t.Errorf("isTrue(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"false", "0", "no", "", "TRUE", "True"} {
		if isTrue(s) {
			t.Errorf("isTrue(%q) = true, want false", s)
		}
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("BURLING_TEST_X", "set")
	if got := envOr("BURLING_TEST_X", "def"); got != "set" {
		t.Errorf("envOr(set) = %q, want set", got)
	}
	if got := envOr("BURLING_TEST_UNSET_Y", "def"); got != "def" {
		t.Errorf("envOr(unset) = %q, want def", got)
	}
}
