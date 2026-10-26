# session

包路径：`mml-server/internal/session`

`Handle(ctx, conn, mapper)` 完成单连接读写循环。`readRequestLine` 识别 LF/CRLF、剥离 IAC、执行 4096 字节上限。

测试：`internal/session/session_test.go`。
