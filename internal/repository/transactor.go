package repository

import (
	"context"
	"errors"
)

var ErrTransactionRequired = errors.New(
	"transaction is required",
)

type Transactor interface {
	WithinTx(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}
