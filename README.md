# openclaw-go (Game Server Template)

一个可直接扩展的 Golang 游戏服务端基础模板，包含：

- 网关层（TCP 连接管理、会话管理）
- 协议层（统一帧协议编解码）
- 协议路由层（按 `msg_id` 路由到业务处理器）
- 示例业务（`ping` / `echo`）

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

## 快速启动

```bash
go run ./cmd/server -addr :9000
```

默认监听 `:9000`。

## 消息号示例

- `1001`：PingReq
- `1002`：PingResp
- `1003`：EchoReq
- `1004`：EchoResp

## 扩展建议

1. 增加认证链路（登录鉴权 / token 校验）
2. 在 `router` 前后加入中间件（限流、日志、追踪）
3. 协议层支持 Protobuf / FlatBuffers
4. 网关和业务分离为独立进程，通过 RPC 通信
