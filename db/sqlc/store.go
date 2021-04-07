package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides functions to execute the quries and txn
type Store struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTxn executes a function within a database transaction
func (store *Store) execTxn(ctx context.Context, fn func(*Queries) error) error {
	// use default isolation level RC(READ COMITTED)
	tx, err := store.db.BeginTx(ctx, nil /*&sql.TxOptions{}*/)
	if err != nil {
		return err
	}

	q := New(tx)
	// execute the call back function
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf(" txn err: %v, roll back error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// TransferTxnParams contains the input params of the transfer txn
type TransferTxnParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxnResult contains the result of a transfer txn
type TransferTxnResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTxn performs a money transfer from one account to the another
// It 1. creates a transfer record, 2. add account entries, 3. and update accounts' balance whin a single database txn
func (store *Store) TransferTxn(ctx context.Context, arg TransferTxnParams) (TransferTxnResult, error) {
	var result TransferTxnResult

	// passing the callback function
	err := store.execTxn(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount, // negative
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// and update accounts' balance

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}

			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}

			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return result, err
}
