package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTxn(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// concurrency( 并发 )
	n := 3
	amount := int64(10)

	fmt.Println(">> origin:", account1.Balance, account2.Balance)

	// channels connects the go routines.
	errs := make(chan error)
	results := make(chan TransferTxnResult)

	for i := 0; i < n; i++ {
		txnName := fmt.Sprintf("txn %v", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txnKey, txnName)
			result, err := store.TransferTxn(ctx, TransferTxnParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}
	// check the results
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// check is the transfer really in the DB
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, account1.ID)
		require.Equal(t, fromEntry.Amount, -amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toEntry.AccountID, account2.ID)
		require.Equal(t, toEntry.Amount, amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)
		// check accounts' balance

		fmt.Println(">> txn", fromAccount.Balance, toAccount.Balance)

		diff1 := account1.Balance - fromAccount.Balance // account1 is the state of the fromAccount before the transfer. so diff1 must be positive
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		// in this test func. we make (n = 5) txns, and every txn transfer (amount = 10) from acount1 to acount2.
		// so the the diff1 must be 1 * amount, 2 * amount, 3 * amount, 4 * amount...
		require.True(t, diff1%10 == 0)

		// check which round of this transfer.
		k := int(diff1 / amount)
		// round time must be [1, n]
		require.True(t, k >= 1 && k <= n)
		// every round must run only ones, we make a map to record the runned round
		require.NotContains(t, existed, k)
		existed[k] = true
	}
	// after all txns, check the updated account
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	fmt.Println(">> after", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)
}
