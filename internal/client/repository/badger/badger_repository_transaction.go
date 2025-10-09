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
	"github.com/symbioticfi/relay/pkg/tracing"
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

func (r *Repository) doUpdateInTxWithLock(ctx context.Context, name string, f func(ctx context.Context) error, lockMap *sync.Map, key any) error {
	ctx, span := tracing.StartDBSpan(ctx, "update", name)
	defer span.End()

	mutexInterface, ok := lockMap.Load(key)
	if !ok {
		newMutex := &mutexWithUseTime{}
		newMutex.lastAccessNs.Store(time.Now().UnixNano())
		mutexInterface, _ = lockMap.LoadOrStore(key, newMutex)
	}
	activeMutex := mutexInterface.(*mutexWithUseTime)

	tracing.AddEvent(span, "acquiring_lock")
	activeMutex.lock()
	defer activeMutex.unlock()
	tracing.AddEvent(span, "lock_acquired")

	err := r.doUpdateInTx(ctx, name, f)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		tracing.RecordError(span, err)
	}
	return err
}

func (r *Repository) doUpdateInTx(ctx context.Context, name string, f func(ctx context.Context) error) error {
	ctx, span := tracing.StartDBSpan(ctx, "update", name)
	defer span.End()

	if getTxn(ctx) != nil {
		startSingle := time.Now()

		err := f(ctx)

		prevName := nameFromCtx(ctx)
		queryName := prevName + "/update:" + name
		r.metrics.ObserveRepoQueryDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(startSingle))

		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			tracing.RecordError(span, err)
		}
		return err
	}

	queryName := "update:" + name
	start := time.Now()

	tracing.AddEvent(span, "starting_transaction")
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
		tracing.AddEvent(span, "transaction_conflict")
		err = errors.Errorf("transaction conflict: %w", entity.ErrTxConflict)
	} else if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		err = errors.Errorf("failed to do update in tx: %w", err)
	}

	r.metrics.ObserveRepoQueryDuration(queryName, status, time.Since(start))
	r.metrics.ObserveRepoQueryTotalDuration(queryName, status, time.Since(start))

	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		tracing.RecordError(span, err)
	} else {
		tracing.AddEvent(span, "transaction_committed")
	}

	return err
}

func (r *Repository) doViewInTx(ctx context.Context, name string, f func(ctx context.Context) error) error {
	ctx, span := tracing.StartDBSpan(ctx, "view", name)
	defer span.End()

	if getTxn(ctx) != nil {
		startSingle := time.Now()

		err := f(ctx)

		prevName := nameFromCtx(ctx)
		queryName := prevName + "/view:" + name
		r.metrics.ObserveRepoQueryDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(startSingle))

		if err != nil {
			tracing.RecordError(span, err)
		}
		return err
	}
	start := time.Now()

	queryName := "view:" + name

	tracing.AddEvent(span, "starting_view_transaction")
	err := r.db.View(func(txn *badger.Txn) error {
		txnCtx := r.withName(
			context.WithValue(ctx, badgerTxnKey, txn),
			queryName,
		)
		return f(txnCtx)
	})

	r.metrics.ObserveRepoQueryDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(start))
	r.metrics.ObserveRepoQueryTotalDuration(queryName, lo.Ternary(err == nil, "ok", "error"), time.Since(start))

	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to do view in tx: %w", err)
	}

	tracing.AddEvent(span, "view_transaction_completed")
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
