## 发生场景


## 调试方法
- txnname: 利用上下文context.WithValue() 给协程传入事务名，可以获取发生死锁的事务。