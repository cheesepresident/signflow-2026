package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"signflow-2026/internal/auth"
	"signflow-2026/internal/db"
	"signflow-2026/internal/email"
)

const signerLinkTTL = 30 * 24 * time.Hour

func (h *Handlers) InviteSigners(w http.ResponseWriter, r *http.Request) {
	user := mustUser(r)
	id, ok := parseUUID(chiURLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	doc, err := h.Queries.GetDocumentForOwner(r.Context(), db.GetDocumentForOwnerParams{ID: id, OwnerID: user.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, "get document", err)
		return
	}
	if doc.Status != "draft" {
		http.Error(w, "signers can only be invited while the document is a draft", http.StatusUnprocessableEntity)
		return
	}

	emails := parseEmails(r.FormValue("emails"))
	if len(emails) == 0 {
		h.renderDocumentDetail(w, r, doc, "Enter at least one valid email address.")
		return
	}

	type invite struct{ email, link string }
	var invites []invite

	tx, err := h.Pool.Begin(r.Context())
	if err != nil {
		h.serverError(w, r, "begin tx", err)
		return
	}
	defer tx.Rollback(context.Background())
	q := h.Queries.WithTx(tx)

	for _, addr := range emails {
		raw, hash, err := auth.GenerateToken()
		if err != nil {
			h.serverError(w, r, "generate token", err)
			return
		}
		if _, err := q.CreateSigner(r.Context(), db.CreateSignerParams{
			DocumentID: doc.ID,
			Email:      addr,
			TokenHash:  hash,
			ExpiresAt:  pgtype.Timestamptz{Time: time.Now().Add(signerLinkTTL), Valid: true},
		}); err != nil {
			h.serverError(w, r, "create signer", err)
			return
		}
		invites = append(invites, invite{email: addr, link: h.Cfg.BaseURL + "/sign/" + raw})
	}

	if _, err := q.MarkDocumentSent(r.Context(), db.MarkDocumentSentParams{ID: doc.ID, OwnerID: user.ID}); err != nil {
		h.serverError(w, r, "mark sent", err)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		h.serverError(w, r, "commit", err)
		return
	}

	for _, inv := range invites {
		text := fmt.Sprintf(
			"You've been invited to sign \"%s\" on SignFlow.\n\n"+
				"Open your signing link (valid 30 days):\n%s\n",
			doc.Filename, inv.link)
		if err := h.Mailer.Send(r.Context(), email.Message{
			To:      inv.email,
			Subject: fmt.Sprintf("Please sign \"%s\"", doc.Filename),
			Text:    text,
			HTML:    fmt.Sprintf(`<p>You've been invited to sign <strong>%s</strong>.</p><p><a href="%s">Open your signing link</a> (valid 30 days).</p>`, doc.Filename, inv.link),
		}); err != nil {
			h.Log.Error("send signer invite", "err", err, "to", inv.email)
		}
	}
	h.Log.Info("document sent for signature", "id", doc.ID.String(), "signers", len(invites))

	http.Redirect(w, r, "/documents/"+doc.ID.String(), http.StatusSeeOther)
}

func parseEmails(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r ==
			'\n' || r ==
			'\r' || r ==
			',' || r ==
			';' || r ==
			' ' || r ==
			'\t'

	})
	seen := map[string]bool{}
	var out []string
	for _, f := range fields {
		addr := normalizeEmail(f)
		if addr == "" || seen[addr] || !validEmail(addr) {
			continue
		}
		seen[addr] = true
		out = append(out, addr)
	}
	return out
}
