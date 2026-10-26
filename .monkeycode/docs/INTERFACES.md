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

## Mapping File

缺省路径：`configs/mapping.txt`

```text
# request<TAB>response
PING	PONG
HELLO	WORLD
```
