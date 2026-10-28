# trea-one 文档

## 项目概述

仓库同时包含 mml-server 与已拆出的 NRF 网元。

mml-server 是基于 TELNET/TCP 的字符串应答服务。客户端按行发送请求，服务端按映射文件精确匹配并回写应答。

NRF 是独立 Go module（`nfs/nrf`），以独立进程提供 Nnrf_NFManagement 与 Nnrf_NFDiscovery。

**目标用户**: 运维与联调人员

**核心价值**:
- 标准 telnet 或裸 TCP 即可接入
- 映射文件调整应答，无需改代码
- 多会话并发，单连接故障不影响其他会话

## 技术选型

| 类别 | 选择 | 理由 |
|------|------|------|
| 语言 | Go 1.22+ | 标准库即可完成 TCP 并发与信号处理 |
| 协议 | TCP 行协议 + IAC 剥离 | 兼容 telnet 与裸 TCP |
| 配置 | TAB 分隔文本文件 | 启动时一次性加载 |

## 文档导航

- 当前工作区 `.monkeycode/docs/ARCHITECTURE.md` - 系统架构
- 当前工作区 `.monkeycode/docs/INTERFACES.md` - CLI 与内部接口
- 当前工作区 `.monkeycode/docs/DEVELOPER_GUIDE.md` - 运行与测试
- 当前工作区 `.monkeycode/docs/专有概念/MappingTable.md` - 映射表
- 当前工作区 `.monkeycode/docs/专有概念/Session.md` - 会话与行协议
- 当前工作区 `.monkeycode/docs/专有概念/NFProfile.md` - NRF NF Profile
- 当前工作区 `.monkeycode/docs/模块/mapping.md` - mapping 模块
- 当前工作区 `.monkeycode/docs/模块/session.md` - session 模块
- 当前工作区 `.monkeycode/docs/模块/server.md` - server 模块
- 当前工作区 `.monkeycode/docs/模块/nrf.md` - NRF 网元
