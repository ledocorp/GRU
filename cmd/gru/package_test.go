package main

import (
	"strings"
	"testing"

	"github.com/ledocorp/gru/internal/version"
)

func TestReleaseLDFlagsQuotesProductName(t *testing.T) {
	flags := releaseLDFlags()
	if !strings.Contains(flags, `-X "github.com/ledocorp/gru/internal/version.Product=`+version.Product+`"`) {
		t.Fatalf("expected quoted Product in ldflags, got: %q", flags)
	}
	if strings.Contains(flags, "Product=Gru Notepad") && !strings.Contains(flags, `"`) {
		t.Fatalf("unquoted space in Product ldflags: %q", flags)
	}
}
