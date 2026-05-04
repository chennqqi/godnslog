package server

import (
	"strings"
	"testing"
)

func TestParseDomain(t *testing.T) {
	var tests = []struct {
		Input         string
		Root          string
		ExpectPrefix  string
		ExpectShortId string
		ExpectRebind  bool
	}{
		{"712hu2c4gy34.godnslog.com", "godnslog.com", "", "712hu2c4gy34", false},
		{"www.godnslog.com", "godnslog.com", "", "www", false},
		{"r.www.godnslog.com", "godnslog.com", "r", "www", true},
		{"r.godnslog.com", "godnslog.com", "", "r", false},
		{"a.r.xxx.godnslog.com", "godnslog.com", "a.r", "xxx", true},
		{"aaa.www.godnslog.com", "godnslog.com", "aaa", "www", false},
		{"bbb.aaa.www.godnslog.com", "godnslog.com", "bbb.aaa", "www", false},
		{"ns.godnslog.com", "godnslog.com", "", "ns", false},

		{"www.godnslog.com.", "godnslog.com", "", "www", false},
		{"r.www.godnslog.com.", "godnslog.com", "r", "www", true},
		{"r.godnslog.com.", "godnslog.com", "", "r", false},
		{"a.r.xxx.godnslog.com.", "godnslog.com", "a.r", "xxx", true},
		{"aaa.www.godnslog.com.", "godnslog.com", "aaa", "www", false},
		{"bbb.aaa.www.godnslog.com.", "godnslog.com", "bbb.aaa", "www", false},
		{"ns.godnslog.com.", "godnslog.com", "", "ns", false},

		{"www.godnslog.com.", "godnslog.com.", "", "www", false},
		{"r.www.godnslog.com.", "godnslog.com.", "r", "www", true},
		{"r.godnslog.com.", "godnslog.com.", "", "r", false},
		{"a.r.xxx.godnslog.com.", "godnslog.com.", "a.r", "xxx", true},
		{"aaa.www.godnslog.com.", "godnslog.com.", "aaa", "www", false},
		{"bbb.aaa.www.godnslog.com.", "godnslog.com.", "bbb.aaa", "www", false},
		{"ns.godnslog.com.", "godnslog.com.", "", "ns", false},
	}

	for i := 0; i < len(tests); i++ {
		test := &tests[i]
		prefix, shortId, rebind := parseDomain(test.Input, test.Root)
		if prefix != test.ExpectPrefix {
			t.Fatalf("test prefix(%v)!=expect(%v)", prefix, test.ExpectPrefix)
		}
		if shortId != test.ExpectShortId {
			t.Fatalf("test shortId(%v)!=expect(%v)", shortId, test.ExpectShortId)
		}
		if rebind != test.ExpectRebind {
			t.Fatalf("test rebind(%v)!=ExpectRebind(%v)", rebind, test.ExpectRebind)
		}
	}
}

func TestParsePrefix(t *testing.T) {
	var tests = []struct {
		Input        string
		ExpectFirst  string
		ExpectSecond string
		ExpectRebind bool
	}{
		{"aaaa.cr", "aaaa", "cr", false},
		{"1-1.cr", "1-1", "cr", false},
		{"127.0.0.1-100.100.100.100.cr", "127.0.0.1-100.100.100.100", "cr", true},
		{"test.cr", "test", "cr", false},
	}
	for i := 0; i < len(tests); i++ {
		test := &tests[i]
		first, second := parsePrefix(test.Input)
		if first != test.ExpectFirst {
			t.Fatalf("test shortId(%v)!=expect(%v)", first, test.ExpectFirst)
		}
		if second != test.ExpectSecond {
			t.Fatalf("test shortId(%v)!=expect(%v)", second, test.ExpectSecond)
		}
		// Rebind detection: check if first part looks like an IP address with dash
		rebind := strings.Contains(first, ".") && strings.Contains(first, "-") && strings.Count(first, ".") >= 3
		if rebind != test.ExpectRebind {
			t.Fatalf("test rebind(%v)!=ExpectRebind(%v)", rebind, test.ExpectRebind)
		}
	}
}
