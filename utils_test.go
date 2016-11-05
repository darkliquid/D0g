package main

import (
	"strings"
	"testing"
)

func TestGetUIDFromMention(t *testing.T) {
	const mention1 = "<@134214314314>"
	const mention2 = "<@!134214314314>"
	const expected = "134214314314"

	if s := getUIDFromMention(mention1); s != expected {
		t.Errorf("expected %q got %q", expected, s)
	}

	if s := getUIDFromMention(mention2); s != expected {
		t.Errorf("expected %q got %q", expected, s)
	}
}

func TestGetGuildIDFromMessage(t *testing.T) {
	t.Skip("can't test this as it requires being connected to a discord server")
}

func TestCleanDiscordString(t *testing.T) {
	const test1 = "   `test\ntest\ntest \\`\\\\\\\\\\\\"
	const expected1 = "test test test"
	var test2 = strings.Repeat(test1, 10)
	const expected2 = "test test test \\\\\\\\\\\\\\   test test test \\\\\\\\\\\\\\   test test test \\\\\\\\\\\\\\   test test test..."

	if s := cleanDiscordString(test1); s != expected1 {
		t.Errorf("expected %q got %q", expected1, s)
	}

	if s := cleanDiscordString(test2); s != expected2 {
		t.Errorf("expected %q got %q", expected2, s)
	}
}
