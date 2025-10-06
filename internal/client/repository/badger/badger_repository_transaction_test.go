package badger

import (
	"context"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_DoUpdateInTx(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("Success - Simple Operation", func(t *testing.T) {
		var executedInTx bool

		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			// Verify we have a transaction in the context
			txn := getTxn(ctx)
			require.NotNil(t, txn, "Transaction should be available in context")

			// Perform a simple write operation
			err := txn.Set([]byte("test_key"), []byte("test_value"))
			require.NoError(t, err)

			executedInTx = true
			return nil
		})

		require.NoError(t, err)
		assert.True(t, executedInTx, "Function should have been executed")

		// Verify the data was actually committed
		err = repo.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("test_key"))
			if err != nil {
				return err
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			assert.Equal(t, "test_value", string(value))
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Rollback on Error", func(t *testing.T) {
		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			txn := getTxn(ctx)
			require.NotNil(t, txn)

			// Write some data
			err := txn.Set([]byte("rollback_key"), []byte("rollback_value"))
			require.NoError(t, err)

			// Return an error to trigger rollback
			return errors.New("intentional error for rollback test")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "intentional error")

		// Verify the data was not committed due to rollback
		err = repo.db.View(func(txn *badger.Txn) error {
			_, err := txn.Get([]byte("rollback_key"))
			return err
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, badger.ErrKeyNotFound))
	})

	t.Run("Nested Transactions - Same Transaction Context", func(t *testing.T) {
		var outerTxn, innerTxn *badger.Txn

		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			outerTxn = getTxn(ctx)
			require.NotNil(t, outerTxn)

			// Nested transaction should reuse the same transaction
			return repo.doUpdateInTx(ctx, "", func(ctx context.Context) error {
				innerTxn = getTxn(ctx)
				require.NotNil(t, innerTxn)

				// Should be the same transaction object
				assert.Equal(t, outerTxn, innerTxn, "Nested transaction should reuse the same transaction")

				// Write data using the nested context
				err := innerTxn.Set([]byte("nested_key"), []byte("nested_value"))
				require.NoError(t, err)

				return nil
			})
		})

		require.NoError(t, err)

		// Verify the nested write was committed
		err = repo.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("nested_key"))
			if err != nil {
				return err
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			assert.Equal(t, "nested_value", string(value))
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Multiple Operations in Same Transaction", func(t *testing.T) {
		testData := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			txn := getTxn(ctx)
			require.NotNil(t, txn)

			// Write multiple keys in the same transaction
			for key, value := range testData {
				err := txn.Set([]byte(key), []byte(value))
				if err != nil {
					return err
				}
			}

			return nil
		})

		require.NoError(t, err)

		// Verify all data was committed atomically
		err = repo.db.View(func(txn *badger.Txn) error {
			for key, expectedValue := range testData {
				item, err := txn.Get([]byte(key))
				if err != nil {
					return err
				}

				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}

				assert.Equal(t, expectedValue, string(value), "Value mismatch for key: %s", key)
			}
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Update Operation in View Transaction - Should Fail", func(t *testing.T) {
		err := repo.doViewInTx(context.Background(), "", func(ctx context.Context) error {
			txn := getTxn(ctx)
			require.NotNil(t, txn)

			// Attempt to perform write operation in view transaction
			// This should fail because view transactions are read-only
			err := txn.Set([]byte("write_in_view_key"), []byte("write_in_view_value"))
			return err
		})

		require.Error(t, err)
		// BadgerDB returns ErrReadOnlyTxn when trying to write in a read-only transaction
		assert.Contains(t, err.Error(), "read-only")
	})

	t.Run("Delete Operation in View Transaction - Should Fail", func(t *testing.T) {
		// First create a key to attempt to delete
		testKey := []byte("delete_test_key")
		err := repo.db.Update(func(txn *badger.Txn) error {
			return txn.Set(testKey, []byte("delete_test_value"))
		})
		require.NoError(t, err)

		// Now try to delete it in a view transaction
		err = repo.doViewInTx(context.Background(), "", func(ctx context.Context) error {
			txn := getTxn(ctx)
			require.NotNil(t, txn)

			// Attempt to delete in view transaction
			// This should fail because view transactions are read-only
			err := txn.Delete(testKey)
			return err
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "read-only")

		// Verify the key still exists (delete didn't work)
		err = repo.db.View(func(txn *badger.Txn) error {
			_, err := txn.Get(testKey)
			return err
		})
		require.NoError(t, err, "Key should still exist since delete in view transaction should have failed")
	})
}
