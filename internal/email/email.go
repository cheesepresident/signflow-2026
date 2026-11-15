package email

import (
	"context"
	"log/slog"
	"strings"
)

type Message struct {
	To      string
	Subject string

	HTML string
	Text string
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

func New(kind, apiKey, from string, log *slog.Logger) Sender {
	if strings.EqualFold(kind, "resend") {
		log.Info("email: using Resend sender", "from", from)
		return NewResendSender(apiKey, from)
	}
	log.Info("email: using console sender (links print to stdout, no email is sent)")
	return NewConsoleSender(nil)
}
