package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"signflow-2026/internal/db"
	"signflow-2026/internal/web"
	"strings"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := mustUser(r)
	docs, err := h.Queries.ListDocumentsByOwner(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "list documents", err)
		return
	}
	render(w, r, http.StatusOK, web.Dashboard(h.nav(r), viewDocuments(docs)))
}

func (h *Handlers) UploadDocument(w http.ResponseWriter, r *http.Request) {
	user := mustUser(r)

	r.Body = http.MaxBytesReader(w, r.Body, h.Cfg.MaxUploadBytes+1024)

	file, header, err := r.FormFile("file")
	if err != nil {
		// Covers both "no file" and "too large" (MaxBytesReader error).
		h.Log.Warn("upload rejected", "err", err, "user", user.Email)
		http.Error(w, "upload failed: file missing or too large", http.StatusRequestEntityTooLarge)
		return
	}
	defer file.Close()

	if header.Size == 0 {
		http.Error(w, "upload failed: empty file", http.StatusBadRequest)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	key, size, hash, err := h.Store.Save(r.Context(), file)
	if err != nil {
		h.serverError(w, r, "store file", err)
		return
	}

	doc, err := h.Queries.CreateDocument(r.Context(), db.CreateDocumentParams{
		OwnerID:     user.ID,
		Filename:    sanitizeFilename(header.Filename),
		ContentType: contentType,
		Size:        size,
		FileHash:    hash,
		StorageKey:  key,
	})
	if err != nil {
		_ = h.Store.Delete(r.Context(), key) // don't orphan the blob
		h.serverError(w, r, "create document", err)
		return
	}
	h.Log.Info("document uploaded", "id", doc.ID.String(), "owner", user.Email, "size", size, "sha256", hash)

	h.respondWithDocList(w, r, user)
}
func (h *Handlers) ViewDocument(w http.ResponseWriter, r *http.Request) {
	doc, ok := h.ownedDocument(w, r)
	if !ok {
		return
	}
	h.renderDocumentDetail(w, r, doc, "")
}
func (h *Handlers) renderDocumentDetail(w http.ResponseWriter, r *http.Request, doc db.Document, errMsg string) {
	signers, err := h.Queries.ListSignersByDocument(r.Context(), doc.ID)
	if err != nil {
		h.serverError(w, r, "list signers", err)
		return
	}
	render(w, r, http.StatusOK, web.DocumentDetail(
		h.nav(r), viewDocument(doc), viewSigners(signers, doc), errMsg,
	))
}

func (h *Handlers) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	doc, ok := h.ownedDocument(w, r)
	if !ok {
		return
	}
	rc, err := h.Store.Open(r.Context(), doc.StorageKey)
	if err != nil {
		h.serverError(w, r, "open file", err)
		return
	}
	defer rc.Close()

	safe := sanitizeFilename(doc.Filename)
	w.Header().Set("Content-Type", doc.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", doc.Size))
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", safe, url.PathEscape(doc.Filename)))
	if _, err := io.Copy(w, rc); err != nil {
		h.Log.Warn("stream download", "err", err, "id", doc.ID.String())
	}
}

func (h *Handlers) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	user := mustUser(r)
	id, ok := parseUUID(chiURLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	doc, err := h.Queries.GetDocumentForOwner(r.Context(), db.GetDocumentForOwnerParams{ID: id, OwnerID: user.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r) // not owned / not found — don't leak which
		return
	}
	if err != nil {
		h.serverError(w, r, "get document", err)
		return
	}
	if doc.Status != "draft" {
		http.Error(w, "only draft documents can be deleted", http.StatusUnprocessableEntity)
		return
	}

	n, err := h.Queries.DeleteDraftDocument(r.Context(), db.DeleteDraftDocumentParams{ID: id, OwnerID: user.ID})
	if err != nil {
		h.serverError(w, r, "delete document", err)
		return
	}
	if n == 1 {
		// Best-effort blob cleanup; a leftover file is harmless.
		_ = h.Store.Delete(r.Context(), doc.StorageKey)
		h.Log.Info("document deleted", "id", doc.ID.String(), "owner", user.Email)
	}

	h.respondWithDocList(w, r, user)
}

func (h *Handlers) ownedDocument(w http.ResponseWriter, r *http.Request) (db.Document, bool) {
	user := mustUser(r)
	id, ok := parseUUID(chiURLParam(r, "id"))
	if !ok {
		http.NotFound(w, r)
		return db.Document{}, false
	}
	doc, err := h.Queries.GetDocumentForOwner(r.Context(), db.GetDocumentForOwnerParams{ID: id, OwnerID: user.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return db.Document{}, false
	}
	if err != nil {
		h.serverError(w, r, "get document", err)
		return db.Document{}, false
	}
	return doc, true
}

func (h *Handlers) respondWithDocList(w http.ResponseWriter, r *http.Request, user db.User) {
	if r.Header.Get("HX-Request") == "" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	docs, err := h.Queries.ListDocumentsByOwner(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "list documents", err)
		return
	}
	render(w, r, http.StatusOK, web.DocumentList(viewDocuments(docs)))
}

func mustUser(r *http.Request) db.User {
	u, _ := UserFrom(r.Context())
	return u
}

func parseUUID(s string) (pgtype.UUID, bool) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return u, false
	}
	return u, u.Valid
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	name = strings.Map(func(rn rune) rune {
		if rn < 0x20 || rn == 0x7f || rn == '"' {
			return -1
		}
		return rn
	}, name)
	name = strings.TrimSpace(name)
	if name == "" {
		return "document"
	}
	return name
}
