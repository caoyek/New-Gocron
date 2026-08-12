package service

import (
	"strings"
	"testing"
)

func TestNormalizeWhitelist(t *testing.T) {
	list, err := NormalizeWhitelist("192.168.1.10\n192.168.1.0/24, 2001:db8::1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 whitelist entries, got %d", len(list))
	}
	if !IPAllowed("192.168.1.88", list) {
		t.Fatal("expected IPv4 CIDR match")
	}
	if !IPAllowed("2001:db8::1", list) {
		t.Fatal("expected IPv6 match")
	}
	if IPAllowed("10.0.0.1", list) {
		t.Fatal("unexpected whitelist match")
	}
}

func TestNormalizeWhitelistRejectsInvalidEntry(t *testing.T) {
	if _, err := NormalizeWhitelist("192.168.1.10\ninvalid-ip"); err == nil {
		t.Fatal("expected invalid whitelist entry error")
	}
}

func TestNormalizeUsernameLimitsStoredValue(t *testing.T) {
	username := "  " + strings.Repeat("用", 70) + "  "
	normalized := normalizeUsername(username)
	if got := len([]rune(normalized)); got != 64 {
		t.Fatalf("expected 64 characters, got %d", got)
	}
}
