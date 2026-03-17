# openclaw-go (Game Server Template)

一个可直接扩展的 Golang 游戏服务端基础模板，包含：

- 网关层（TCP 连接管理、会话管理）
- 协议层（统一帧协议编解码）
- 协议路由层（按 `msg_id` 路由到业务处理器）
- MySQL 持久化（玩家角色信息）
- 示例业务（`ping` / `echo` / `create character` / `get character`）

## 架构分层

```
client <-> gateway(server/session) <-> protocol(frame/json) <-> router <-> handlers
```

目录结构：

```
cmd/server/main.go          # 服务入口
internal/gateway/           # 网关：监听、连接、会话、收发
internal/protocol/          # 协议：帧结构、JSON 序列化
internal/router/            # 路由：msg_id -> handler
internal/handlers/          # 示例业务处理器
internal/messages/ids.go    # 消息号定义
internal/player/            # 玩家领域模型与存储接口
internal/storage/mysql/     # MySQL 存储实现
scripts/mysql/init.sql      # MySQL 初始化脚本
docker-compose.yml          # 本地 MySQL 容器
```

## 协议格式（TCP 二进制帧）

Header 固定 6 字节（大端序）：

- `payload_len`：4 字节，无符号整型，仅表示 payload 长度
- `msg_id`：2 字节，无符号整型

Body：

- `payload`：业务内容（默认 JSON）

即：

```
+----------------+-------------+------------------+
| payload_len(4) | msg_id(2)   | payload(N bytes) |
+----------------+-------------+------------------+
```

## 快速启动（MySQL + Server）

### 1) 启动 MySQL

```bash
docker compose up -d mysql
```

默认数据库参数：

- host: `127.0.0.1`
- port: `3306`
- database: `gserver`
- user: `gserver`
- password: `gserver`

### 2) 启动服务端

```bash
go run ./cmd/server -addr :9000 -mysql-dsn "gserver:gserver@tcp(127.0.0.1:3306)/gserver?charset=utf8mb4&parseTime=true&loc=Local"
```

默认监听 `:9000`。

## 消息号示例

- `1001`：PingReq
- `1002`：PingResp
- `1003`：EchoReq
- `1004`：EchoResp
- `2001`：CreateCharacterReq
- `2002`：CreateCharacterResp
- `2003`：GetCharacterReq
- `2004`：GetCharacterResp

## 扩展建议

1. 增加认证链路（登录鉴权 / token 校验）
2. 在 `router` 前后加入中间件（限流、日志、追踪）
3. 协议层支持 Protobuf / FlatBuffers
4. 网关和业务分离为独立进程，通过 RPC 通信
