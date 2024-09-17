# LiteRedis

[English](README.md)

LiteRedis 是一个轻量级的、类 Redis 的内存数据存储系统，使用 Go 语言实现。它旨在提供一个简单但功能强大的键值存储解决方案，支持多种数据结构和操作。

<p align="center">
  <img src="https://img.shields.io/badge/Python-snow?logo=python&logoColor=3776AB" alt="" />
  <img src="https://img.shields.io/badge/Java-snow?logo=coffeescript&logoColor=FC4C02" alt="" />
  <img src="https://img.shields.io/badge/Go-snow?logo=go&logoColor=00ADD8" alt="" />
  <img src="https://img.shields.io/badge/Rust-snow?logo=rust&logoColor=000000" alt="" />
  <img src="https://img.shields.io/badge/TypeScript-snow?logo=typescript&logoColor=3178C6" alt="" />
</p>

## 功能特性

- 支持字符串、列表、哈希表和集合等数据结构
- 提供与 Redis 兼容的命令接口
- 支持键过期功能
- 内存存储，快速读写
- 并发安全
- 支持多个数据库

## 快速开始

### 安装

确保已经安装了 Go（版本 1.16 或更高）。
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

### 贡献

欢迎贡献代码、报告问题或提出新功能建议。请遵循以下步骤：

1. Fork 项目
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 待办事项

- 实现更多的 Redis 命令
- 添加持久化功能
- 实现主从复制
- 添加事务支持
- 优化内存使用


[↑ top](#literedis)
<br><br><br><br><hr>