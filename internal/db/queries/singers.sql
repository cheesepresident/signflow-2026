-- name: CreateSigner :one
INSERT INTO signers (document_id, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: ListSignersByDocument :many
SELECT * FROM signers WHERE document_id = $1 ORDER BY created_at;