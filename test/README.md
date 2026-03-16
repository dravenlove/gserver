# Test Tools

这个目录用于放置所有测试相关内容（联调工具 + 单元测试）。

## Unit Tests

- `test/protocol/frame_test.go`
- `test/router/router_test.go`

运行：

```bash
go test ./...
```

## TCP Client

路径：`test/client/main.go`

### 1) 启动服务端

```bash
go run ./cmd/server -addr :9000 -mysql-dsn "gserver:gserver@tcp(127.0.0.1:3306)/gserver?charset=utf8mb4&parseTime=true&loc=Local"
```

### 2) 发送 ping

```bash
go run ./test/client -addr 127.0.0.1:9000 -mode ping -count 3
```

### 3) 发送 echo

```bash
go run ./test/client -addr 127.0.0.1:9000 -mode echo -text "hello gateway"
```

### 4) 交互模式

```bash
go run ./test/client -addr 127.0.0.1:9000 -mode interactive
```

交互命令：

- 普通文本：按 `echo` 发送
- `/ping`：发送一次 ping
- `/create <player_id> <name>`：创建角色
- `/get <player_id>`：查询角色
- `/quit`：退出

### 5) 创建角色

```bash
go run ./test/client -addr 127.0.0.1:9000 -mode create-player -player-id player_001 -name "claw"
```

### 6) 查询角色

```bash
go run ./test/client -addr 127.0.0.1:9000 -mode get-player -player-id player_001
```
