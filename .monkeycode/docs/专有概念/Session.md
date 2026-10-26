# Session

Session 是一条已建立的 TCP/TELNET 连接。Accept 后在独立 goroutine 中运行 `session.Handle`。

行为：
- 按字节扫描，`\n` 为行结束；若前一字节为 `\r` 则剥离
- 读到 `0xFF` 时消费 TELNET IAC 协商字节，不纳入命令
- 累计超过 4096 字节仍无换行则关闭连接
- 应答固定为 `{response}\r\n`
- 对端关闭、读/写错误时结束该 Session，其他 Session 继续
