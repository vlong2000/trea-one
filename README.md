# trea-one

mml-server 已实现，`go test ./...` 全部通过。

NRF 已拆为独立 Go module（`nfs/nrf`），独立进程，NF Profile 存于内存。

```bash
# 构建 NRF
make nrf

# 单独运行 NRF
./bin/nrf -c config/nrfcfg.yaml
```

默认监听 `http://127.0.0.1:8000`，提供 `/nnrf-nfm/v1`、`/nnrf-disc/v1`、`/bootstrapping`。

启动：

```bash
go run ./cmd/mml-server --host 127.0.0.1 --port 2323 --mapping configs/mapping.txt
```

再用 `telnet 127.0.0.1 2323` 发送 `PING` 应得 `PONG`。

默认监听 `0.0.0.0:2323`，映射文件 `configs/mapping.txt`（TAB 分隔）；未匹配回 `ERROR: unknown command`。
