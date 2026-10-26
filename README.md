# trea-one

mml-server 已实现，`go test ./...` 全部通过。

启动：

```bash
go run ./cmd/mml-server --host 127.0.0.1 --port 2323 --mapping configs/mapping.txt
```

再用 `telnet 127.0.0.1 2323` 发送 `PING` 应得 `PONG`。

默认监听 `0.0.0.0:2323`，映射文件 `configs/mapping.txt`（TAB 分隔）；未匹配回 `ERROR: unknown command`。
