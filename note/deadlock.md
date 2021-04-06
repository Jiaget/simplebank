
# 发生场景 与 调试方法
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
开启多个终端，按照打印日志的顺序进行并发事务的模拟，直到出现死锁为止。
 
测试过程中，发现：当txn 3执行到`get account 1`时， txn 3 挂起等待，txn1 执行 `get account 1`时，pg报错
```
ERROR:  deadlock detected
DETAIL:  Process 969 waits for ShareLock on transaction 1452; blocked by process 986.
Process 986 waits for ShareLock on transaction 1453; blocked by process 969.
HINT:  See server log for query details.
CONTEXT:  while locking tuple (0,1) in relation "accounts"
```
# 分析与解决方法

在该情况下，可以在wiki 中找到pg lock相关的信息。
获取查询pg后台记录的关于锁的信息的SQL语句。这里就不赘述了。

我们可以查询到`get account 1`是被 `create Entry 1`。`get account 1`使用的是`SELECT FOR UPDATE` 的查询语句，可以视为写操作，但是这两个操作作用于两张不同的表，为什么会出现阻塞？

答案是：`外键约束`

虽然两个写操作作用于两张不同的表，它们被外键约束联系在了一起。当我们把外键约束去掉之后，再运行测试代码，发现死锁消失了。但是实际上我们不能没有这些约束，因为这些约束能保证数据的一致性。（`entry`表记录了`account`表的主键，没有约束，`entry`表中的`account_id` 可能出现问题。）

pg之所以会在我们对entry表操作时给`account`表上锁，就是因为担心我们会修改`account`表的主键，而`account`的主键会影响`entry`表，因为他们之间有外键约束。当然我们的操作并不会影响主键，我们需要告诉pg这点:
```
-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;
```

最后再运行测试代码，ok。

但是，死锁问题仍然没有解决。。。

# 死锁的残存

