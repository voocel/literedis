# LiteRedis

[English](README.md)

LiteRedis 是一个用 Go 语言实现的轻量级 Redis-like 内存存储系统。它支持字符串、哈希、列表和集合等数据类型，并提供了一个简单的网络服务器来处理客户端请求。

## 功能特性

- 支持的数据类型：
  - 字符串（String）
  - 哈希（Hash）
  - 列表（List）
  - 集合（Set）
- 支持基本的 Redis 命令
- 多数据库支持
- 键过期功能
- 简单的网络服务器

## 快速开始

### 安装

确保已经安装了 Go（版本 1.16 或更高）。然后，克隆此仓库：
```
git clone https://github.com/yourusername/literedis.git
cd literedis
```

### 构建
```
go build -o literedis ./cmd/server
```

在项目根目录下运行以下命令来构建项目：

### 运行

构建完成后，可以通过以下命令启动服务器：
```
./literedis
```

默认情况下，服务器将在 localhost:6379 上监听连接。

## 使用方法

可以使用任何 Redis 客户端连接到 LiteRedis 服务器。例如，使用 `redis-cli`：
```
redis-cli -p 6379
```

然后，可以执行支持的命令，如：
```
SET mykey "Hello, LiteRedis!"
GET mykey
HSET myhash field1 "value1"
HGET myhash field1
LRANGE mylist 0 -1
```

## 支持的命令

LiteRedis 支持以下命令：

- 字符串操作：SET, GET, DEL, EXISTS, EXPIRE, TTL
- 哈希操作：HSET, HGET, HDEL, HLEN
- 列表操作：LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE
- 集合操作：SADD, SMEMBERS, SREM, SCARD
- 通用操作：SELECT, FLUSHDB, FLUSHALL

## 测试
```
go test ./...
```

[↑ top](#literedis)
<br><br><br><br><hr>