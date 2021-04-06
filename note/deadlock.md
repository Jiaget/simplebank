## 发生场景


## 调试方法
- txnname: 利用上下文context.WithValue() 给协程传入事务名，可以获取发生死锁的事务。
- 打Log之后再运行测试程序可以看到：
```
  >> origin: 662 656
txn 3  : create transfer
txn 3  : create Entry 1
txn 1  : create transfer
txn 3  : create Entry 2
txn 2  : create transfer
txn 1  : create Entry 1
txn 3  : get account 1
txn 1  : create Entry 2
txn 2  : create Entry 1
txn 1  : get account 1
txn 2  : create Entry 2
txn 2  : get account 1
--- FAIL: TestTransferTxn (1.19s)
    c:\Users\ASUS\go\simpleBank\db\sqlc\store_test.go:44: 
        	Error Trace:	store_test.go:44
        	Error:      	Received unexpected error:
        	            	pq: deadlock detected
        	Test:       	TestTransferTxn
```

一条transferTxn的SQL逻辑如下
```
BEGIN;

INSERT INTO transfers (from_account_id, to_account_id, amount) VALUES (1, 2, 10) RETURNING *;

INSERT INTO entries (account_id, amount) VALUES (1, -10) RETURNING *;
INSERT INTO entries (account_id, amount) VALUES (2, 10) RETURNING *;

SELECT * FROM accounts WHERE id = 1 FOR UPDATE;
UPDATE accounts SET balance = 90 WHERE id = 1 RETURNING *;

SELECT * FROM accounts WHERE id = 2 FOR UPDATE;
UPDATE accounts SET balance = 110 WHERE id = 2 RETURNING *;

ROLLBACK;
```
开启多个终端，按照打印日志的顺序进行并发事务的模拟，直到出现死锁为止