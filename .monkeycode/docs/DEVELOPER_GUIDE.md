# 开发指南

## 项目目的

mml-server 提供 TELNET 接入与字符串精确应答，用于联调与测试场景按配置回写固定响应。

**核心职责**:
- 在指定地址接受 TCP/TELNET 连接
- 按行匹配映射表并回写应答
- 多会话并发与进程信号退出

## 环境搭建

### 前置条件

- Go >= 1.22

### 运行

```bash
go run ./cmd/mml-server --host 127.0.0.1 --port 2323 --mapping configs/mapping.txt
```

连接：

```bash
telnet 127.0.0.1 2323
```

输入 `PING` 应收到 `PONG`。

### 测试

```bash
go test ./...
```

测试覆盖 mapping 加载、session 行协议、server 并发与端口占用、CLI 非法参数。

## 项目结构说明

- `cmd/mml-server`：CLI 与信号
- `internal/mapping`：映射表
- `internal/session`：单连接协议
- `internal/server`：监听与生命周期
- `configs/mapping.txt`：缺省映射

## 开发规范

- 仅使用 Go 标准库，不引入第三方依赖
- 映射表启动后只读，Session 无共享可变状态
- 日志使用标准库 `log`，Session 日志带对端地址

## 常见任务

新增应答：在 Mapping File 增加一行 `request<TAB>response` 后重启进程。

调整监听：使用 `--host` 与 `--port`。
