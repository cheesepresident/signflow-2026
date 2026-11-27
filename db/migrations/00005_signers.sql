-- +goose Up
-- +goose StatementBegin
CREATE TABLE signers (
                         id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                         document_id     UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
                         email           TEXT        NOT NULL,
                         token_hash      TEXT        NOT NULL UNIQUE,
                         status          TEXT        NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending', 'signed')),
                         signed_name     TEXT,
                         signed_at       TIMESTAMPTZ,
                         signed_doc_hash TEXT,
                         expires_at      TIMESTAMPTZ NOT NULL,
                         created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Can't invite the same email to the same document twice.
                         UNIQUE (document_id, email)
);

CREATE INDEX idx_signers_document_id ON signers(document_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE signers;
-- +goose StatementEnd