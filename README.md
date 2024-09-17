# LiteRedis

[中文版](README_CN.md)

LiteRedis is a lightweight Redis-like in-memory storage system implemented in Go. It supports various data types such as strings, hashes, lists, and sets, and provides a simple network server to handle client requests.

<p align="center">
  <img src="https://img.shields.io/badge/Python-snow?logo=python&logoColor=3776AB" alt="" />
  <img src="https://img.shields.io/badge/Java-snow?logo=coffeescript&logoColor=FC4C02" alt="" />
  <img src="https://img.shields.io/badge/Go-snow?logo=go&logoColor=00ADD8" alt="" />
  <img src="https://img.shields.io/badge/Rust-snow?logo=rust&logoColor=000000" alt="" />
  <img src="https://img.shields.io/badge/TypeScript-snow?logo=typescript&logoColor=3178C6" alt="" />
</p>

## Features

- Supported data types:
  - String
  - Hash
  - List
  - Set
- Basic Redis commands
- Multiple database support
- Key expiration
- Simple network server

## Quick Start

### Installation

Ensure you have Go installed (version 1.16 or higher). Then, clone this repository:
```
git clone https://github.com/yourusername/literedis.git
cd literedis
```


### Build

Run the following command in the project root directory to build the project:
```
go build -o literedis ./cmd/server
```

### Run

After building, you can start the server with the following command:
```
./literedis
```

By default, the server will listen on localhost:6379.

## Usage

You can use any Redis client to connect to LiteRedis. For example, using `redis-cli`:
```
redis-cli -p 6379
```

Then, you can execute supported commands, such as:
```
SET mykey "Hello, LiteRedis!"
GET mykey
HSET myhash field1 "value1"
HGET myhash field1
LRANGE mylist 0 -1

## Supported Commands

LiteRedis supports the following commands:

- String operations: SET, GET, DEL, EXISTS, EXPIRE, TTL
- Hash operations: HSET, HGET, HDEL, HLEN
- List operations: LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE
- Set operations: SADD, SMEMBERS, SREM, SCARD
- General operations: SELECT, FLUSHDB, FLUSHALL

## Testing

To run all tests:
```
go test ./...
```