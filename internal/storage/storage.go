package storage

import (
	"context"
	"io"
)

type Store interface {
	Save(ctx context.Context, r io.Reader) (key string, size int64, sha256hex string, err error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
