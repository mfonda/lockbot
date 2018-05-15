package main

import (
	"github.com/mfonda/slash"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input           string
		expectedCommand string
		expectedNotes   string
	}{
		{"foo bar", "foo", "bar"},
		{"foo", "foo", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		cmd, notes := parseCommand(test.input)
		if cmd != test.expectedCommand {
			t.Errorf("command incorrect: expected %s, got %s", test.expectedCommand, cmd)
		}
		if notes != test.expectedNotes {
			t.Errorf("notes incorrect: expected %s, got %s", test.expectedNotes, notes)
		}
	}
}

func TestReply(t *testing.T) {
	text := "foo"
	r, err := reply(text)
	if err != nil {
		t.Error("Unexpected error")
	}
	if r.ResponseType != "in_channel" {
		t.Errorf("Expected in_channel response, got %s", r.ResponseType)
	}
	if r.Text != text {
		t.Errorf("Expected response text '%s', got %s", text, r.Text)
	}
}

func TestLock(t *testing.T) {
	req := newRequest("username", "foo bar")
	resp, _ := lockHandler(req)
	l, _ := locks["foo"]
	if l.String() != resp.Text {
		t.Errorf("Failed to lock: expected %s, got %s", l.String(), resp.Text)
	}
}

func TestUnlock(t *testing.T) {
	req := newRequest("username", "foo")
	locks["foo"] = lock{}
	unlockHandler(req)
	_, exists := locks["foo"]
	if exists {
		t.Errorf("Lock should have been removed, but still exists: %s", locks)
	}
}

func newRequest(username string, text string) *slash.Request {
	r := &slash.Request{}
	r.UserName = username
	r.Text = text
	return r
}
