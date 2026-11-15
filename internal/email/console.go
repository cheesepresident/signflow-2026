package email

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type ConsoleSender struct {
	w  io.Writer
	mu sync.Mutex // serialize writes so concurrent emails don't interleave
}

func NewConsoleSender(w io.Writer) *ConsoleSender {
	if w == nil {
		w = os.Stdout
	}
	return &ConsoleSender{w: w}
}

func (s *ConsoleSender) Send(_ context.Context, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	const bar = "──────────────────────────────────────────────────────────────"
	fmt.Fprintf(s.w, "\n┌%s\n", bar)
	fmt.Fprintf(s.w, "│ 📧  DEV EMAIL (not actually sent)\n")
	fmt.Fprintf(s.w, "│ To:      %s\n", msg.To)
	fmt.Fprintf(s.w, "│ Subject: %s\n", msg.Subject)
	fmt.Fprintf(s.w, "├%s\n", bar)
	for _, line := range strings.Split(strings.TrimRight(msg.Text, "\n"), "\n") {
		fmt.Fprintf(s.w, "│ %s\n", line)
	}
	fmt.Fprintf(s.w, "└%s\n\n", bar)
	return nil
}
