# 需求实施计划

- [x] 1. 设置项目结构
  - 初始化 `go.mod`（模块路径 `mml-server`）
  - 创建 `cmd/mml-server`、`internal/server`、`internal/session`、`internal/mapping`、`configs`
  - 添加缺省 Mapping File `configs/mapping.txt`（含 `PING`/`HELLO` 示例）

- [x] 2. 实现映射表加载与查询
  - [x] 2.1 实现 `internal/mapping`：`Load`、`Lookup`
    - 支持 `--mapping` 缺省路径 `configs/mapping.txt`
    - 忽略空行与 `#` 注释；`request<TAB>response`；重复 key 后者覆盖
    - 文件缺失或零条合法映射时返回 error
    - 覆盖 Requirement 4、Correctness Property 1
  - [x] 2.2 为 mapping 编写单元测试
    - 合法文件、缺失文件、空文件、重复 key、无 TAB 跳过、注释/空行、Lookup 命中/未命中

- [x] 3. 实现 Session 行协议
  - [x] 3.1 实现 `internal/session`：按行读取、IAC 剥离、超限关闭、CRLF 写回
    - LF/CRLF 得到同一 Request Line；应答 `{response}\r\n`
    - 超 4096 字节关闭 Session；EOF/网络错误关闭 Session
    - 覆盖 Requirement 2、5 与 Correctness Property 3/4/5
  - [x] 3.2 为 session 编写单元测试
    - `\n`、`\r\n`、连续两行、超长行、IAC 协商字节、未匹配应答

- [x] 4. 实现 TCP 监听与多会话
  - [x] 4.1 实现 `internal/server`：`ListenAndServe`、并发 Session、优雅退出
    - 地址占用返回 error；ctx 取消后停止 Accept、关闭连接，等待上限 3 秒
    - 覆盖 Requirement 1、3、5.4
  - [x] 4.2 为 server 编写集成测试
    - PING 应答、CRLF 与 LF、未知命令、双客户端并发、一端断开、端口占用

- [x] 5. 实现进程入口
  - [x] 5.1 实现 `cmd/mml-server/main.go`
    - 解析 `--host`/`--port`/`--mapping`；缺省 `0.0.0.0:2323` 与 `configs/mapping.txt`
    - `--port` 非法、mapping/listen 失败退出码 1；SIGINT/SIGTERM 退出码 0
    - 覆盖 Requirement 1.2/1.3/1.6、4.1/4.2、5.4

- [x] 6. 检查点 - 确保所有测试通过
  - 运行 `go test ./...`，如有疑问请询问用户
