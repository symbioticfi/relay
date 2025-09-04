package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
)

type ctxKey struct{}

var badgerTxnKey ctxKey

func (r *Repository) DoUpdateInTx(ctx context.Context, f func(ctx context.Context) error) error {
	if getTxn(ctx) != nil {
		return f(ctx)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		txnCtx := context.WithValue(ctx, badgerTxnKey, txn)
		return f(txnCtx)
	})
}

func (r *Repository) DoViewInTx(ctx context.Context, f func(ctx context.Context) error) error {
	if getTxn(ctx) != nil {
		return f(ctx)
	}

	return r.db.View(func(txn *badger.Txn) error {
		txnCtx := context.WithValue(ctx, badgerTxnKey, txn)
		return f(txnCtx)
	})
}

func getTxn(ctx context.Context) *badger.Txn {
	if txn, ok := ctx.Value(badgerTxnKey).(*badger.Txn); ok {
		return txn
	}
	return nil
}
