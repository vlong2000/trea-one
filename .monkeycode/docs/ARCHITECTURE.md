# 架构设计

## 概述

mml-server 是一个基于 TELNET/TCP 的字符串应答服务。运维与联调人员通过 IP 与端口接入后，按行发送请求字符串，服务端依据启动时加载的映射表精确匹配并回写应答。

进程使用 Go 标准库实现监听、会话并发与行协议。每条连接在独立 goroutine 中处理，映射表启动后只读，会话之间无共享可变状态。标准 `telnet` 客户端与裸 TCP 客户端均可按行交互。

默认监听 `0.0.0.0:2323`，默认映射文件为 `configs/mapping.txt`。启动失败（映射缺失、端口占用、端口非法）以退出码 1 结束；SIGINT/SIGTERM 后优雅退出，退出码 0。

## 技术栈

**语言与运行时**
- Go 1.22+（当前环境为 Go 1.25）

**框架**
- 无第三方依赖
- 测试使用 Go 标准库 `testing`

**数据存储**
- 文本映射文件（`request<TAB>response`），启动时一次性加载到内存 `map[string]string`

**基础设施**
- 独立进程，TCP Listen/Accept

**外部服务**
- 无

## 项目结构

```
mml-server/
├── cmd/mml-server/       # 进程入口与 CLI
├── internal/mapping/     # 映射表加载与查询
├── internal/session/     # 单连接行协议
├── internal/server/      # TCP 监听与会话生命周期
├── configs/mapping.txt   # 缺省映射文件
└── go.mod
```

**入口点**
- `cmd/mml-server/main.go` - 解析参数、加载映射、注册信号、启动监听

## 子系统

### cmd/mml-server
**目的**: 进程入口与失败退出码
**位置**: `cmd/mml-server/`
**关键文件**: `main.go`
**依赖**: `internal/mapping`、`internal/server`
**被依赖**: 无

### mapping
**目的**: 从文本文件加载精确映射并查询
**位置**: `internal/mapping/`
**关键文件**: `mapping.go`
**依赖**: 无
**被依赖**: `cmd/mml-server`、`internal/server`、`internal/session`

### session
**目的**: 单连接读写循环、IAC 剥离、超限关闭
**位置**: `internal/session/`
**关键文件**: `session.go`
**依赖**: `internal/mapping`
**被依赖**: `internal/server`

### server
**目的**: TCP 监听、并发 Session、优雅退出
**位置**: `internal/server/`
**关键文件**: `server.go`
**依赖**: `internal/mapping`、`internal/session`
**被依赖**: `cmd/mml-server`

## 图表

```mermaid
flowchart LR
    ClientA["TELNET Client A"]
    ClientB["TELNET Client B"]
    Main["cmd/mml-server"]
    Mapper["internal/mapping"]
    Srv["internal/server"]
    Sess["internal/session"]
    File["configs/mapping.txt"]

    File --> Mapper
    Main --> Mapper
    Main --> Srv
    ClientA --> Srv
    ClientB --> Srv
    Srv --> Sess
    Sess --> Mapper
```

```mermaid
sequenceDiagram
    participant Client
    participant Server as internal/server
    participant Session as internal/session
    participant Mapper as internal/mapping

    Client->>Server: TCP connect
    Server->>Session: Handle conn
    Client->>Session: Request Line plus LF or CRLF
    Session->>Mapper: Lookup request
    Mapper-->>Session: response or ERROR unknown command
    Session-->>Client: response plus CRLF
```
