package account

import (
	"context"
	"strings"
	"testing"
)

type fakeSender struct {
	to      string
	subject string
	body    string
	err     error
}

func (f *fakeSender) SendMail(to, subject, body string) error {
	f.to = to
	f.subject = subject
	f.body = body
	return f.err
}

func TestSMTPPasswordResetNotifierSends(t *testing.T) {
	sender := &fakeSender{}
	n := NewSMTPPasswordResetNotifier(SMTPConfig{
		From:    "no-reply@example.com",
		BaseURL: "https://app.example.com",
	}, sender)

	err := n.SendPasswordReset(context.Background(), "user@example.com", "token123")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if sender.to != "user@example.com" {
		t.Fatalf("unexpected to: %s", sender.to)
	}
	if sender.subject == "" {
		t.Fatalf("expected subject")
	}
	if !strings.Contains(sender.body, "https://app.example.com/reset-password?token=token123") {
		t.Fatalf("expected reset link in body")
	}
}

func TestSMTPPasswordResetNotifierRequiresBaseURL(t *testing.T) {
	n := NewSMTPPasswordResetNotifier(SMTPConfig{From: "no-reply@example.com"}, &fakeSender{})
	err := n.SendPasswordReset(context.Background(), "user@example.com", "token123")
	if err == nil {
		t.Fatalf("expected error for missing base URL")
	}
}
