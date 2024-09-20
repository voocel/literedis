# LiteRedis API 文档

LiteRedis 是一个轻量级的 Redis 类内存存储系统，支持多种数据类型和基本的 Redis 命令。以下是 LiteRedis 支持的主要命令及其用法。

## 目录
- [字符串操作](#字符串操作)
  - [SET](#set)
  - [GET](#get)
  - [DEL](#del)
  - [EXISTS](#exists)
  - [EXPIRE](#expire)
  - [TTL](#ttl)
- [哈希操作](#哈希操作)
  - [HSET](#hset)
  - [HGET](#hget)
  - [HDEL](#hdel)
  - [HLEN](#hlen)
  - [HKEYS](#hkeys)
  - [HVALS](#hvals)
  - [HGETALL](#hgetall)
- [列表操作](#列表操作)
  - [LPUSH](#lpush)
  - [RPUSH](#rpush)
  - [LPOP](#lpop)
  - [RPOP](#rpop)
  - [LLEN](#llen)
  - [LRANGE](#lrange)
- [集合操作](#集合操作)
  - [SADD](#sadd)
  - [SMEMBERS](#smembers)
  - [SREM](#srem)
  - [SCARD](#scard)
- [通用操作](#通用操作)
  - [SELECT](#select)
  - [FLUSHDB](#flushdb)
  - [FLUSHALL](#flushall)

## 字符操作

### SET
设置键的值。

**语法**:
```
SET key value
```
**示例**:
```
SET mykey "Hello, LiteRedis!"
```


### GET
获取键的值。

**语法**:
```
GET key
```
**示例**:
```
GET mykey
```

### DEL
删除一个或多个键。

**语法**:
```
DEL key [key ...]
```
**示例**:
```
DEL mykey
```

### EXISTS
检查键是否存在。

**语法**:
```
EXISTS key
```
**示例**:
```
EXISTS mykey
```

### EXPIRE
设置键的过期时间。

**语法**:
```
EXPIRE key seconds
```
**示例**:
```
EXPIRE mykey 60
```

### TTL
获取键的剩余生存时间。

**语法**:
```
TTL key
```
**示例**:
```
TTL mykey
```

## 哈希操作

### HSET
设置哈希表中字段的值。

**语法**:
```
HSET key field value [field value ...]
```
**示例**:
```
HSET myhash field1 "value1" field2 "value2"
```

### HGET
获取哈希表中字段的值。

**语法**:
```
HGET key field
```
**示例**:
```
HGET myhash field1
```

### HDEL
删除哈希表中的一个或多个字段。

**语法**:
```
HDEL key field [field ...]
```
**示例**:
```
HDEL myhash field1
```

### HLEN
获取哈希表中字段的数量。

**语法**:
```
HLEN key
```
**示例**:
```
HLEN myhash
```

### HKEYS
获取哈希表中所有字段名。

**语法**:
```
HKEYS key
```
**示例**:
```
HKEYS myhash
```

### HVALS
获取哈希表中所有字段值。

**语法**:
```
HVALS key
```
**示例**:
```
HVALS myhash
```

### HGETALL
获取哈希表中所有字段和值。

**语法**:
```
HGETALL key
```
**示例**:
```
HGETALL myhash
```

## 列表操作

### LPUSH
将一个或多个值插入到列表头部。

**语法**:
```
LPUSH key value [value ...]
```
**示例**:
```
LPUSH mylist "value1" "value2"
```

### RPUSH
将一个或多个值插入到列表尾部。

**语法**:
```
RPUSH key value [value ...]
```
**示例**:
```
RPUSH mylist "value1" "value2"
```

### LPOP
移除并返回列表的第一个元素。

**语法**:
```
LPOP key
```
**示例**:
```
LPOP mylist
```

### RPOP
移除并返回列表的最后一个元素。

**语法**:
```
RPOP key
```
**示例**:
```
RPOP mylist
```

### LLEN
获取列表的长度。

**语法**:
```
RPOP key
```
**示例**:
```
RPOP mylist
```