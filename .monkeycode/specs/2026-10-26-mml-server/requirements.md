# Requirements Document

## Introduction

mml-server 是一个基于 TELNET 的字符串应答服务。客户端通过 IP 与端口建立连接后，服务端按配置文件中的精确映射规则，将收到的字符串转换为对应的应答字符串并回写给客户端。本需求覆盖连接建立、会话收发、映射匹配与异常处理。

## Glossary

- **System**: mml-server，提供 TELNET 接入与字符串应答能力的服务进程
- **Client**: 通过 IP 与端口接入 System 的 TELNET 客户端
- **Session**: Client 与 System 之间的一条已建立的 TELNET 连接
- **Request Line**: Client 发送的一条完整输入字符串，行结束符为 LF 或 CRLF
- **Response Line**: System 根据映射规则生成、回写给 Client 的应答字符串，以 CRLF 结束
- **Mapping Table**: 将 Request Line 精确映射到 Response Line 的配置集合
- **Mapping File**: 存放 Mapping Table 的文本配置文件，每行一条映射，格式为 `request<TAB>response`
- **Listen Address**: System 监听的 IP 与端口组合，由启动参数 `--host` 与 `--port` 指定，缺省为 `0.0.0.0:2323`

## Requirements

### Requirement 1: TELNET 接入

**User Story:** AS 运维人员, I want 通过 IP 与端口建立 TELNET 连接, so that 我可以用标准 TELNET 客户端访问应答服务

#### Acceptance Criteria

1. WHEN System 启动完成, the System SHALL 在 Listen Address 上接受 TELNET 连接
2. WHEN 启动参数提供 `--host` 与 `--port`, the System SHALL 使用上述参数组成 Listen Address
3. WHEN 启动参数未提供 `--host` 或 `--port`, the System SHALL 对缺失项使用缺省值 `0.0.0.0` 或 `2323`
4. WHEN Client 使用 TELNET 协议连接到 Listen Address, the System SHALL 建立一条独立 Session 并保持该 Session 直到 Client 断开或 System 主动关闭
5. WHILE Session 处于已连接状态, the System SHALL 允许 Client 发送 Request Line 并接收 Response Line
6. IF Listen Address 被占用, the System SHALL 以非零退出码终止启动并输出可定位的错误信息

### Requirement 2: 字符串收发与精确映射应答

**User Story:** AS 客户端使用者, I want 发送字符串后收到对应应答, so that 我可以按约定完成命令交互

#### Acceptance Criteria

1. WHEN Session 收到一条以 LF 或 CRLF 结束的输入, the System SHALL 去掉行结束符后得到 Request Line
2. WHEN Session 得到 Request Line, the System SHALL 在 Mapping Table 中按整行完全相等查找映射项
3. WHEN Request Line 与 Mapping Table 中某一项的请求字符串完全相等, the System SHALL 将该项对应的 Response Line 写回同一 Session
4. IF Request Line 在 Mapping Table 中无完全相等的映射项, the System SHALL 向同一 Session 写回固定未匹配应答 `ERROR: unknown command`
5. WHEN System 写回 Response Line, the System SHALL 以 CRLF 结束该应答，使 Client 能按行读取完整结果

### Requirement 3: 多会话并发

**User Story:** AS 运维人员, I want 多个客户端同时连接, so that 测试与联调可以并行进行

#### Acceptance Criteria

1. WHILE 已有 Session 处于连接状态, the System SHALL 继续接受新的 Client 连接
2. WHEN 多条 Session 同时存在, the System SHALL 按 Session 独立完成 Request Line 接收与 Response Line 回写
3. WHEN 某一 Session 断开, the System SHALL 释放该 Session 占用的资源，并保持其余 Session 继续工作

### Requirement 4: 映射配置

**User Story:** AS 运维人员, I want 通过配置文件维护请求与应答的精确对应关系, so that 无需改代码即可调整应答内容

#### Acceptance Criteria

1. WHEN System 启动且启动参数提供 `--mapping`, the System SHALL 从该路径加载 Mapping File
2. WHEN System 启动且启动参数未提供 `--mapping`, the System SHALL 从缺省路径 `configs/mapping.txt` 加载 Mapping File
3. WHEN Mapping File 中某一非空、非注释行同时包含以 TAB 分隔的请求字符串与应答字符串, the System SHALL 将该项纳入 Mapping Table
4. WHEN Mapping Table 中存在重复的请求字符串, the System SHALL 保留文件中后出现的映射项
5. IF Mapping File 缺失或无法按行解析为合法映射, the System SHALL 以非零退出码终止启动并输出可定位的错误信息

### Requirement 5: 连接与输入异常

**User Story:** AS 客户端使用者, I want 连接异常时服务保持可用, so that 单个客户端故障不影响整体服务

#### Acceptance Criteria

1. WHEN Client 主动关闭 Session, the System SHALL 结束该 Session 并释放对应资源
2. IF Session 在读取 Request Line 过程中发生网络错误, the System SHALL 关闭该 Session 并记录错误原因
3. IF Request Line 去掉行结束符后的长度超过 4096 字节, the System SHALL 关闭该 Session 并记录超限原因
4. WHEN System 收到进程终止信号 SIGINT 或 SIGTERM, the System SHALL 停止接受新连接、关闭已有 Session 并以零退出码退出
