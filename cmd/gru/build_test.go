package main

import "testing"

func TestParseAppPlatform(t *testing.T) {
	cases := []struct {
		args     []string
		wantApp  string
		wantPlat string
	}{
		{[]string{"hello"}, "hello", hostPlatform()},
		{[]string{"demo", "windows"}, "demo", "windows"},
		{[]string{"notepad", "linux"}, "notepad", "linux"},
		{[]string{"windows"}, "notepad", "windows"},
	}
	for _, tc := range cases {
		app, plat, rest, err := parseAppPlatform(tc.args)
		if err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if app != tc.wantApp || plat != tc.wantPlat {
			t.Fatalf("%v: got %s/%s want %s/%s", tc.args, app, plat, tc.wantApp, tc.wantPlat)
		}
		if len(rest) != 0 && tc.args[0] != "hello" {
			// ok if flags remain; empty for these cases
		}
		_ = rest
	}
}

func TestProductLDFlagsQuotesHello(t *testing.T) {
	flags := productLDFlags(products["hello"])
	if !containsQuotedProduct(flags, "Hello Gru") {
		t.Fatalf("expected quoted Hello Gru in ldflags, got %q", flags)
	}
}

func containsQuotedProduct(flags, product string) bool {
	needle := `-X "github.com/ledocorp/gru/internal/version.Product=` + product + `"`
	return len(flags) > 0 && (stringIndex(flags, needle) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
