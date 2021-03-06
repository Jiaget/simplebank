## 0. 一些准备
使用win10操作系统，因此需要做一些工具准备。
- mingw：轻量级GNU工具，主要目的：在windows使用`make`命令。
- 在powershell 下使用代理。`$Env:http_proxy="http://127.0.0.1:7890";$Env:https_proxy="http://127.0.0.1:7890"`
- 修改go的镜像。否则go get基本无法使用，除非使用了代理。 `go env -w GOPROXY=https://goproxy.cn`
- docker 。 在windows中可以模拟linux环境，有些工具需要编译，可以在docker中编译，或者有直接的docker镜像。
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
    - 当遇到新的需求，需要增加数据库的表/列时，不要移除原有的，增加新的迁移脚本即可。
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

## 7. 事务
相关代码：store.go

Queries 结构体只对一张表进行一个操作，因此Queries结构体不能支持事务。因此另外建一个store结构体

  - Why
    - 1.保证数据可靠与一致。特别是考虑到系统出现故障的情况
    - 2.保证访问数据库的隔离性
      - ACID（原子性、一致性、隔离性、持久性）
  - 并发: 
    - 使用channel 来联系主进程和协程。协程内往channel传送信息，协程外接受channel中的信息。
    - 当两个事务同时访问同一个数据的时候，由于数据库的隔离性为‘读已提交’，当事务1对数据进行了修改，事务2仍然获取该数据的旧数据。这会给转账的业务逻辑带来问题。
      - 解决方案：行锁 `SELECT FOR UPDATE`
        - 数据库的事务层的设计中，‘读已提交’虽然可以提高数据库数据处理的效率，但是，会出现上面的问题。上行锁其实就是提高局部数据的隔离等级，保证当前可能会出现问题的数据的安全性。

- TDD:
  - 测试驱动开发，先编写测试文件使程序出错，再编写程序来通过测试）
- 处理死锁问题
  - 详情见 ./deadlock.md
## 8. CI
github自动化测试部署。

脚本写在`.github/workflows/ci.yml`

更多详情查看文档即可

## 9.RESTful HTTP API
- 一些流行的web框架
  - Gin
  - Beego
  - Echo
  - Revel
  - Martini
  - Fiber
  - Buffalo
- 流行的HTTP路由
  - FastHttp
  - Gorilla Mux
  - HttpRouter
  - Chi

这里使用`Gin`框架
- 验证机制。文档与实例

https://pkg.go.dev/github.com/go-playground/validator
```
Currency string `json:"currency" binding:"require,oneof=RMB USD EUR"`
```

### 自定义验证

使用`validator` 包 自定义验证函数。并在`Gin`中的`binding` tag中注册，定义验证tag名即可
```
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validateCurrency)
	}
```

## 10.从文件或者环境变量加载配置 （viper）
- viper 的功能
  - 找到并加载配置文件
    - JSON
    - TOML
    - YAML
    - ENV
    - INI
  - 从环境变量或者`flags`中读取配置信息
    - 覆盖已有值；
    - 设置默认值。    - 
  - 从远程系统读取配置信息
    - Etcd, consul
  - 远程监控，修改配置文件
- 安装
- `go get github.com/spf13/viper`
- 数据绑定
```
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}
```

## 11.mock
- 优势
  - 独立测试，避免冲突
  - 更快测试，减少连接数据库的时间花费
  - 100%的覆盖，更容易写一些边界的cases（比如errors）
- 方法
  - fake DB:将数据存储在内存里
  - DB stubs:GOMOCK
  
这里我们使用`gomock`来实现` go install github.com/golang/mock/mockgen@v1.5.0`

当前的数据库连接参数写在`store`结构体里,为了方便扩展功能，将`store`改写成接口。

但是会出现的问题是`Queries`结构体中的所有方法都需要填入该接口里。这会消耗大量时间，且会增加整个代码的耦合度。`sqlc`可以实现接口代码的生成，在sqlc.yaml中设置即可。最后
`sqlc generate`

使用`gomock`可以自动起服务并进行测试，具体参考代码在`/api/account_test.go`中

## 12. token认证
- JWT
- PASETO
这两类库函数详细介绍在`token.md`

本项目用两个库各实现了对称密钥算法。两个均可以使用。代码在`token`包。

对于不同程度的数据操作（`Create`, `Get`, `List`, `Transfer`），需要制定不同的授权。授权相关代码，写在`Gin`框架封装的`middleware`中间件中，可以在接受`request`后进行认证后，再提供指定的服务。


## 13. 为项目构建docker镜像
项目根目录下新建文件`Dockerfile`
- 未优化
  - 挑选golang的镜像作为基础镜像。如果希望生成镜像小一点，选择`alpine`版本即可。`FROM golang:1.16.3-alpine3.13`
  - 指定docker镜像内的工作路径`WORKDIR /app`
  - 将文件拷贝进目标路径 `COPY . .`
    - 第一个点代表从当前目录拷贝所有文件
    - 第二个点代表镜像的路径，即镜像内的`/app`路径。
  - 编译获取二进制文件 `RUN go build -o main main.go`
  - 给定端口`EXPOSE 8080`
  - 设定启动镜像时的运行命令`CMD ["/app/main]`
  - 最后编译成docker镜像文件 `docker build -t simplebank:latest .` 最后的点代表Dockfile的路径
虽然最后编译成功，但是通过`docker images`发现生成的镜像文件400多M，非常庞大。接下来需要将这个镜像缩小。
- 优化
  - 镜像文件体积庞大的原因是，镜像不仅包含了二进制文件，还包含了源码。解决方法只需要将镜像中的源码清除即可。
```
# Build stage
FROM golang:1.16-alpine3.13 AS buiilder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["/app/main"]
```
-注意： build 阶段，go build 会下载go.mod中的依赖，由于docker容器无法通过主机的代理来下载依赖，可以在`Dockerfile`中设置go的环境变量，使用国内镜像下载依赖

`RUN go env -w GOPROXY=https://goproxy.cn`

## 14.启动docker镜像
- `docker run --name simplebank -p 8080:8080 simplebank:latest`
  - 启动时报错“无法加载conifg文件，文件找不到”
  - 原因：为了缩小镜像大小，只移动了编译的二进制文件，只需要在`Dockerfile`添加移动配置文件的命令即可。
- 修复了配置文件问题，继续启动docker镜像会发现启动了`GIN`框架的 DEBUG 模式，
```
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
```
  - 按照提示信息，切换到`release`模式 `docker run --name simplebank -p 8080:8080 -e GIN_MODE=release simplebank:latest`
- 在POSTMan 发送请求，出现500的报错。原因是无法连接到 `127.0.0.1:5432` 。这是因为启动的docker镜像服务连接的ip地址不再是主机的IP，而是docker容器里的ip。我们只需要将访问数据库的ip参数修改成同样是docker中pg的IP即可。
  - 查看docker容器的参数 `docker container inspect postgres13` 获取Pg的ip地址
    - 方法一:将连接数据库参数添加到docker 的 `-e` Tag中`docker run --name simplebank -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@172.17.0.3:5432/simple_bank?sslmode=disable" simplebank:latest`
    - 方法二（推荐）使用docker 的 network
      - `docker network ls` 查看docker 当前的networks
      - `docker network inspect bridge` 查看bridge的详细信息。我们可以发现，容器运行在该network下，我们可以自己建一个类似bridge的 network。 这样pg 和simplebank 两个容器可以使用名字相互获取对方的ip地址，ip地址之后再发生变化也不需要修改命令参数了
      - `docker network create bank-network` 创建一个network bank-network。
      - `docker network connect bank-network postgres13`。 将pg连接在该network下。
      - `docker run --name simplebank --network bank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://root:secret@postgres13:5432/simple_bank?sslmode=disable" simplebank:latest` success!