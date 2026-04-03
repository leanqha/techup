package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"strings"

	"techup/internal/apperrors"
)

type EmailSender interface {
	SendMail(to, subject, body string) error
}

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	UseTLS     bool
	SkipVerify bool
	BaseURL    string
}

type SMTPPasswordResetNotifier struct {
	cfg    SMTPConfig
	sender EmailSender
}

func NewSMTPPasswordResetNotifier(cfg SMTPConfig, sender EmailSender) *SMTPPasswordResetNotifier {
	if sender == nil {
		sender = &smtpSender{cfg: cfg}
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &SMTPPasswordResetNotifier{cfg: cfg, sender: sender}
}

func (n *SMTPPasswordResetNotifier) SendPasswordReset(_ context.Context, email, token string) error {
	if n.cfg.BaseURL == "" {
		return apperrors.InvalidArgument("APP_BASE_URL is required for SMTP notifier")
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", n.cfg.BaseURL, url.QueryEscape(token))
	subject, body := buildResetEmail(n.cfg.From, email, resetURL)
	return n.sender.SendMail(email, subject, body)
}

type smtpSender struct {
	cfg SMTPConfig
}

func (s *smtpSender) SendMail(to, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))
	msg := buildMessage(s.cfg.From, to, subject, body)

	if s.cfg.UseTLS {
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			ServerName:         s.cfg.Host,
			InsecureSkipVerify: s.cfg.SkipVerify,
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := s.authIfNeeded(client); err != nil {
			return err
		}
		if err := s.sendMessage(client, s.cfg.From, to, msg); err != nil {
			return err
		}
		return client.Quit()
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host, InsecureSkipVerify: s.cfg.SkipVerify}); err != nil {
			return err
		}
	}
	if err := s.authIfNeeded(client); err != nil {
		return err
	}
	if err := s.sendMessage(client, s.cfg.From, to, msg); err != nil {
		return err
	}
	return client.Quit()
}

func (s *smtpSender) authIfNeeded(client *smtp.Client) error {
	if s.cfg.Username == "" {
		return nil
	}
	return client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host))
}

func (s *smtpSender) sendMessage(client *smtp.Client, from, to string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func buildResetEmail(from, to, resetURL string) (string, string) {
	subject := "Password reset"
	body := fmt.Sprintf(
		"Hello,\n\nTo reset your password, open the link below:\n%s\n\nIf you did not request this, you can ignore this email.\n",
		resetURL,
	)
	return subject, body
}

func buildMessage(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)
}
