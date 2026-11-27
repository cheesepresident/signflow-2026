-- +goose Up
-- +goose StatementBegin
-- Append-only audit trail. Every state change lands exactly one row here.

CREATE TABLE audit_events (
                              seq             BIGINT      GENERATED ALWAYS AS IDENTITY,
                              id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                              document_id     UUID        NOT NULL,
                              event_type      TEXT        NOT NULL,
                              message         TEXT        NOT NULL,
                              actor_type      TEXT        NOT NULL CHECK (actor_type IN ('user', 'signer', 'system', 'anonymous')),
                              actor_user_id   UUID,
                              actor_signer_id UUID,
                              actor_label     TEXT        NOT NULL,
                              created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_events_document ON audit_events (document_id, seq);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION audit_events_append_only() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only: % is not permitted', TG_OP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audit_events_no_update_delete
    BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION audit_events_append_only();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audit_events_no_truncate
    BEFORE TRUNCATE ON audit_events
    FOR EACH STATEMENT EXECUTE FUNCTION audit_events_append_only();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE audit_events;
DROP FUNCTION IF EXISTS audit_events_append_only();
-- +goose StatementEnd