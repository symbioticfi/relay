package badger

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/internal/entity"
)

type ctxTxnKey struct{}

var badgerTxnKey ctxTxnKey

type ctxQueryNameKey struct{}

var ctxQueryName ctxQueryNameKey

// mutexWithUseTime wraps a mutex with a timestamp of last access
type mutexWithUseTime struct {
	mutex        sync.Mutex
	lastAccessNs atomic.Int64 // Unix nanoseconds
}

func (m *mutexWithUseTime) lock() {
	m.mutex.Lock()
	m.lastAccessNs.Store(time.Now().UnixNano())
}

func (m *mutexWithUseTime) unlock() {
	m.mutex.Unlock()
}

func (m *mutexWithUseTime) lastAccess() time.Time {
	return time.Unix(0, m.lastAccessNs.Load())
}

func (m *mutexWithUseTime) tryLock() bool {
	return m.mutex.TryLock()
}

type lock struct {
	lockMap *sync.Map
	key     any
}

func (r *Repository) doUpdateInTxWithLock(ctx context.Context, name string, f func(ctx context.Context) error, locks ...lock) error {
	for _, l := range locks {
		mutexInterface, ok := l.lockMap.Load(l.key)
		if !ok {
			newMutex := &mutexWithUseTime{}
			newMutex.lastAccessNs.Store(time.Now().UnixNano())
			mutexInterface, _ = l.lockMap.LoadOrStore(l.key, newMutex)
		}
		activeMutex := mutexInterface.(*mutexWithUseTime)

		activeMutex.lock()
		defer activeMutex.unlock() //nolint:revive // defer in loop is intentional here
	}

	return r.doUpdateInTx(ctx, name, f)
}

func (r *Repository) doUpdateInTx(ctx context.Context, name string, f func(ctx context.Context) error) error {
	if getTxn(ctx) != nil {
		startSingle := time.Now()

		err := f(ctx)

		prevName := nameFromCtx(ctx)
		queryName := prevName + "/update:" + name
		r.metrics.ObserveRepoQueryDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(startSingle))
		return err
	}

	queryName := "update:" + name
	start := time.Now()
	err := r.db.Update(func(txn *badger.Txn) error {
		txnCtx := r.withName(
			context.WithValue(ctx, badgerTxnKey, txn),
			queryName,
		)
		return f(txnCtx)
	})

	status := lo.Ternary(err == nil, "ok", "error")
	if errors.Is(err, badger.ErrConflict) {
		status = "conflict"
		err = errors.Errorf("transaction conflict: %w", entity.ErrTxConflict)
	} else if err != nil {
		err = errors.Errorf("failed to do update in tx: %w", err)
	}

	r.metrics.ObserveRepoQueryDuration(queryName, status, time.Since(start))
	r.metrics.ObserveRepoQueryTotalDuration(queryName, status, time.Since(start))

	return err
}

func (r *Repository) doViewInTx(ctx context.Context, name string, f func(ctx context.Context) error) error {
	if getTxn(ctx) != nil {
		startSingle := time.Now()

		err := f(ctx)

		prevName := nameFromCtx(ctx)
		queryName := prevName + "/view:" + name
		r.metrics.ObserveRepoQueryDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(startSingle))
		return err
	}
	start := time.Now()

	queryName := "view:" + name
	err := r.db.View(func(txn *badger.Txn) error {
		txnCtx := r.withName(
			context.WithValue(ctx, badgerTxnKey, txn),
			queryName,
		)
		return f(txnCtx)
	})
	if err != nil {
		r.metrics.ObserveRepoQueryDuration(queryName, "error", time.Since(start))
		r.metrics.ObserveRepoQueryTotalDuration(queryName, "error", time.Since(start))

		return errors.Errorf("failed to do view in tx: %w", err)
	}

	r.metrics.ObserveRepoQueryDuration(queryName, "ok", time.Since(start))
	r.metrics.ObserveRepoQueryTotalDuration(queryName, "ok", time.Since(start))

	return nil
}

func getTxn(ctx context.Context) *badger.Txn {
	if txn, ok := ctx.Value(badgerTxnKey).(*badger.Txn); ok {
		return txn
	}
	return nil
}

func (r *Repository) withName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ctxQueryName, name)
}

func nameFromCtx(ctx context.Context) string {
	if name, ok := ctx.Value(ctxQueryName).(string); ok {
		return name
	}
	return "unknown"
}
