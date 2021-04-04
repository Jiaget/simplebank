## 1.设计数据库与SQL生成

https://www.dbdiagram.io

## 2.docker + pg
- 下载docker desktop
- 在 https://hub.docker.com 搜索postgres，找到合适的tag
- 在终端 `docker pull <image>:tag`下拉镜像
- `docker run --name postgres13 POSTGRES_USER=root -e POSTGRES_PASSWORD=mysecretpassword -d postgres:13-alpine` 启动pg镜像
- `docker exec -it postgres13 psql -U root` 运行pg
- `docker logs postgres13` 查看容器日志

## 3. tablePlus
- 将dbdiagram.io生成得sql文件导入tablePlus中运行完成建表工作

## 4. golang-migrate
- 用于数据库架构迁移
- github.com/golang-migrate/migrate -> CLI 选择相应版本下载(下载release版本，使用scoop下载，受网络限制。。。)
- `migrate create -ext sql -dir .\db\migrate\ -seq init_schema` 创建了两个脚本
    - up : 更新最新得schema
    - down : 回退(revert)
- 将SQL 文件中得SQL语句粘至up文件
- 在down文件中写入删除所有表的语句
- `docker exec -it postgres13 /bin/sh` 可以进入docker 的ubuntu shell 界面，执行ubuntu的各种命令
  - `createdb --username=root --owner=root simple_bank`
  - `simple_bank`
- 在docker容器的外面也可以创建db `docker exec -it postgres13 createdb --username=root --owner=root simple_bank`
- `docker exec -it postgres13 psql -U root simple_bank`
- 使用`Makefile`脚本进行自动化操作。
  - 注意事项：
    - windows环境中可以使用gow来模拟linux环境，使用make命令。
    - 在windows写Makefile时，由于tab键和linux有所区别，故可以在vim中写Makefile文件。
#### Makefile 文件
```
postgres:
	docker run --name postgres13 -p 5432:5432 POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:13-alpine

createdb:
	docker exec -it postgres13 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres13 dropdb simple_bank

migrateup:
	migrate -path ./db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path ./db/migrate -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
.PHONY: postgres createdb dropdb migrateup migratedown
```

## 5.生成CRUD代码
- db_sql
  - 快
  - 易出错，且runtime才能捕获问题
- gorm （Golang 的库）
  - CRUD已被实现，所需要的代码少
  - 需要用gorm的函数来写查询代码（新的学习成本）
  - 高负载下运行慢
- sqlx
  - 速度快，易用
  - 同样易出错，且在runtime 才能捕获问题
- sqlc（choose)
  - 快，易用，自动生成代码
  - 编写代码时即可发现SQL的错误
  - 能支持PG但是MySQL只在实验中。。。

~~在 https://github.com/kyleconroy/sqlc 下载安装sqlc。 使用命令`go get github.com/kyleconroy/sqlc/cmd/sqlc`，在此之前需要给终端设置代理~~
```
powershell

$Env:http_proxy="http://127.0.0.1:7890";$Env:https_proxy="http://127.0.0.1:7890"
```
```
CMD

set http_proxy=http://127.0.0.1:7890 & set https_proxy=http://127.0.0.1:7890
```
在windows由于缺少环境，无法直接使用sqlc，这里可以使用docker作为解决方案
  - 拉取镜像 `docker pull kjconroy/sqlc`
  - 运行sqlc `docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc generate`
docker模拟linux环境，基本可以解决此类问题

sqlc配置文件`sqlc.yaml` 相关设置参数参考 `https://docs.sqlc.dev/en/latest/reference/config.html#`
  - `emit_exact_table_names` 若为false .会将数据库的表名单复数化形式命名给结构体 `Table account -> Accounts struct` 结构体名和表名区分开来比较好。

在yaml文件中 `queries` 对应的目录文件下创建SQL文件。写好需要生成的sql语句，运行 `docker run --rm -v C:\Users\ASUS\go\goProject:/src -w /src kjconroy/sqlc generate` 即可在sqlc下生成对应的文件

- sqlc 下的文件中调用其他文件的全局变量有红色波浪线。这是因为项目没有进行初始化module file 运行`go mod init github.com/Jiaget/simplebank` 生成go.mod文件。运行 `go mod tidy` 会帮助下载所有的依赖项。在这里只是处理了项目内部的依赖问题

## 6. 单元测试
获取测试框架 `go get github.com/stretchr/testify` 用来检查测试结果，比if else更方便

获取pg 驱动`go get github.com/lib/pq`

  - 随机生成测试数据 
    - util
      - init 函数： 在每次编译时获取当前时间作为随机数种子。
  - makefile 文件中添加 `go test -v -cover ./...`设置自动化启动测试 `-v` 显示测试的详细命令。 `-cover` 测试的覆盖范围