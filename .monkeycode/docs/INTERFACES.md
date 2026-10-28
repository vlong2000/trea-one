# 接口定义

## CLI

进程入口：`cmd/mml-server`

| 参数 | 缺省 | 约束 |
|------|------|------|
| `--host` | `0.0.0.0` | 监听地址 |
| `--port` | `2323` | 整数 1-65535，非法则退出码 1 |
| `--mapping` | `configs/mapping.txt` | 映射文件路径，缺失或无合法项则退出码 1 |

启动示例：

```bash
go run ./cmd/mml-server --host 127.0.0.1 --port 2323 --mapping configs/mapping.txt
```

退出码：
- `0`：收到 SIGINT/SIGTERM 后正常退出
- `1`：参数错误、映射加载失败、监听失败

## 行协议

客户端通过 TCP（含标准 TELNET）连接 Listen Address。

- Request Line：以 LF 或 CRLF 结束；服务端去掉行结束符后精确匹配
- Response Line：`{response}\r\n`
- 未匹配：`ERROR: unknown command\r\n`
- Request Line 不含行结束符时长度上限 4096 字节，超限关闭该 Session
- TELNET IAC（`0xFF`）及其后续 1 或 2 个协商字节不纳入 Request Line

## mapping.Table

位置：`internal/mapping`

```go
func Load(path string) (*Table, error)
func (t *Table) Lookup(request string) string
```

- `Load`：文件缺失返回 error；空行与 `#` 注释忽略；合法行 `request<TAB>response`；`request` 非空，`response` 可为空；重复 key 后者覆盖；零条合法映射返回 error
- `Lookup`：精确匹配；未命中返回 `"ERROR: unknown command"`

## server.Server

位置：`internal/server`

```go
func New(addr string, mapper *mapping.Table) *Server
func (s *Server) ListenAndServe(ctx context.Context) error
func (s *Server) Addr() string
```

- `ListenAndServe`：绑定 `addr`；`ctx` 取消时关闭 listener；Accept 后为每个连接启动 Session
- 监听失败（含地址占用）返回 error
- 退出等待在途 Session 最多 3 秒，超时后关闭剩余连接

## session.Handle

位置：`internal/session`

```go
func Handle(ctx context.Context, conn net.Conn, mapper *mapping.Table)
```

读取 Request Line、Lookup、以 CRLF 写回。EOF、网络错误、超长行时关闭连接。`ctx` 取消时关闭 `conn`。

## NRF CLI

进程入口：`nfs/nrf/cmd`

| 参数 | 缺省 | 约束 |
|------|------|------|
| `-c` / `--config` | `./config/nrfcfg.yaml` | YAML 路径；缺失或非法则非零退出 |

```bash
make nrf
./bin/nrf -c config/nrfcfg.yaml
```

默认监听 `http://127.0.0.1:8000`。

## NRF SBI

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/bootstrapping` | 返回 OPERATIVE 状态 |
| PUT | `/nnrf-nfm/v1/nf-instances/{id}` | 注册或更新 NF Profile；新建 201，更新 200 |
| GET | `/nnrf-nfm/v1/nf-instances/{id}` | 查询单个 Profile；不存在 404 |
| DELETE | `/nnrf-nfm/v1/nf-instances/{id}` | 注销；成功 204 |
| GET | `/nnrf-nfm/v1/nf-instances` | 列表，可选 `nf-type` |
| GET | `/nnrf-disc/v1/nf-instances` | 发现，必填 `target-nf-type` 与 `requester-nf-type` |

注册体至少包含 `nfType`、`nfStatus`。Profile 存于进程内存，不使用 MongoDB。

## Mapping File

缺省路径：`configs/mapping.txt`

```text
# request<TAB>response
PING	PONG
HELLO	WORLD
```
