# Requirements Document

## Introduction

本需求将 free5GC 核心网按 3GPP 网元（Network Function, NF）拆分为彼此独立的工程与进程。本轮范围限定为核心网元：AMF、SMF、UPF、NRF、AUSF、UDM、UDR、PCF、NSSF。CHF、NEF、BSF、N3IWF、TNGF 与 WebConsole 不在本轮拆分范围内。

本轮 NF 落在同一 Git 仓库内的独立 Go module 目录中，各自拥有独立可执行文件、独立配置与独立生命周期。协议编解码、OpenAPI 模型与公共工具放入独立 Git 仓库中的 Shared Library，由各 NF module 以 Go module 依赖引用。NF 之间仅通过 3GPP 规定的接口通信，运行时不共享同一进程地址空间。

当前工作区尚未包含 free5GC 源码。本需求以 free5GC 官方模块化结构与 3GPP TS 23.501 服务化架构为基准，约束拆分后的工程边界、进程隔离、构建发布与联调方式。

## Glossary

- **System**: 拆分后的 free5GC 核心网工程集合，包含各 NF 独立项目及其联调编排
- **NF**: Network Function，3GPP 定义的 5G 核心网网元
- **In-Scope NF**: 本轮拆分覆盖的网元，包括 AMF、SMF、UPF、NRF、AUSF、UDM、UDR、PCF、NSSF
- **Out-of-Scope NF**: 本轮不拆分的网元，包括 CHF、NEF、BSF、N3IWF、TNGF 以及 WebConsole
- **Control Plane NF**: 本轮范围内通过 SBI（HTTP/2）交互的控制面网元：AMF、SMF、AUSF、UDM、UDR、PCF、NRF、NSSF
- **User Plane NF**: 用户面网元 UPF，通过 N3/N4/N6 与 RAN、SMF、数据网交互
- **NF Project**: 单个 NF 的独立 Go module，位于同一 Git 仓库的独立目录（`nfs/<nf>/`），含独立 `go.mod`、独立 `cmd` 入口、独立配置与独立构建产物
- **NF Process**: 由某一 NF Project 构建出的可执行文件在运行时形成的操作系统进程
- **SBI**: Service-Based Interface，控制面 NF 之间的 HTTP/2 服务化接口
- **NRF**: NF Repository Function，提供 NF 注册与发现
- **Shared Library**: 独立 Git 仓库，仓内包含 `openapi`、`nas`、`ngap`、`pfcp`、`util` 等多个 Go module；被 NF Project 依赖，仅提供库能力，不形成 NF 进程
- **Aggregator Repo**: 承载各 NF Project 目录、编排脚本、配置样例、集成测试与文档的单一 Git 仓库；NF 业务实现只存在于对应 `nfs/<nf>/` 目录
- **Go Workspace**: 仓库根目录的 `go.work`，用于在同一工作区中同时开发多个 NF module
- **In-Process Coupling**: 两个 NF 的业务逻辑在同一操作系统进程内通过函数调用或共享内存交互的耦合方式

## Requirements

### Requirement 1: 按 NF 拆分为独立工程

**User Story:** AS 核心网开发人员, I want 每个 NF 成为独立项目, so that 我可以单独开发、构建、版本化某一个网元

#### Acceptance Criteria

1. WHEN 拆分完成, the System SHALL 在同一 Aggregator Repo 内为下列每个 In-Scope NF 提供一个独立 NF Project 目录：AMF、SMF、UPF、NRF、AUSF、UDM、UDR、PCF、NSSF
2. WHEN 某一 NF Project 存在, the NF Project SHALL 拥有独立的 `go.mod`、独立的源码树与独立的构建入口，且该 module 不将其他 In-Scope NF 的业务包列为构建依赖
3. WHEN 某一 NF Project 构建成功, the NF Project SHALL 产出仅对应该 NF 的可执行文件
4. IF 某一 In-Scope NF 的业务实现位于其他 NF Project 目录或 Aggregator Repo 根目录业务包内, the System SHALL 将该实现迁移到所属 `nfs/<nf>/` 目录
5. WHEN 拆分完成, the System SHALL 保持 Out-of-Scope NF 不进入本轮 `nfs/` 独立工程清单

### Requirement 2: 进程隔离，禁止共享进程

**User Story:** AS 运维人员, I want 每个 NF 运行在独立进程中, so that 单个网元故障、升级或重启不影响其他网元的进程存活

#### Acceptance Criteria

1. WHEN 任一 NF 启动, the System SHALL 以独立 NF Process 运行该 NF，使该进程的地址空间仅包含该 NF 的业务逻辑
2. WHEN 两个不同 NF 需要交互, the System SHALL 仅通过 3GPP 接口完成交互（控制面走 SBI HTTP/2，SMF 与 UPF 走 N4/PFCP，RAN 与 AMF 走 N2/NGAP，RAN 与 UPF 走 N3/GTP-U）
3. WHEN 某一 NF Process 退出, the System SHALL 保持其余已启动 NF Process 继续运行
4. WHEN 运维人员单独启动某一 NF Project 的可执行文件, the NF Process SHALL 在未加载其他 NF 业务代码的前提下完成该 NF 的初始化与对外服务
5. IF 启动编排脚本被调用, the System SHALL 为每个目标 NF 拉起各自的 NF Process，使多个 NF 不以同一进程入口共存

### Requirement 3: 独立配置、生命周期与可观测性

**User Story:** AS 运维人员, I want 每个 NF 使用独立配置并独立启停, so that 我可以按网元调整参数与发布节奏

#### Acceptance Criteria

1. WHEN 某一 NF Process 启动, the NF Process SHALL 从其所属 NF Project 的配置文件（或该 NF 专用的配置路径）读取本 NF 参数
2. WHEN 某一 NF 的配置变更, the System SHALL 允许仅重启该 NF Process 使新配置生效，其余 NF Process 保持运行
3. WHEN 某一 NF Process 收到 SIGINT 或 SIGTERM, the NF Process SHALL 停止本 NF 的对外服务并退出，其余 NF Process 保持运行
4. WHEN 某一 NF Process 启动失败, the System SHALL 以非零退出码结束该 NF Process，并输出可定位到该 NF 的错误信息
5. WHEN 某一 Control Plane NF Process 启动成功, the NF Process SHALL 使用本 NF 的 SBI 绑定地址与端口对外提供服务
6. WHEN 某一 Control Plane NF Process 启动且配置了 NRF 地址, the NF Process SHALL 向 NRF 注册本 NF 的服务信息

### Requirement 4: 共享代码边界

**User Story:** AS 核心网开发人员, I want 协议与公共库可被多个 NF 依赖, so that 拆分后仍能复用编解码与 OpenAPI 模型，同时保持进程隔离

#### Acceptance Criteria

1. WHEN 多个 NF Project 需要同一份协议编解码、OpenAPI 模型或公共工具, the System SHALL 将这些代码放入独立 Git 仓库中的 Shared Library，由各 NF Project 通过 `go.mod` 依赖引用
2. WHEN Shared Library 被引用, the Shared Library SHALL 仅提供库能力，不提供可启动的 NF 进程入口
3. WHEN 某一 NF Project 构建, the NF Project SHALL 允许静态链接所需 Shared Library，运行时仍只产生一个对本 NF 的 NF Process
4. IF 某段代码同时包含某一 NF 的业务流程与可复用协议逻辑, the System SHALL 将业务流程留在该 NF Project，将协议逻辑放入 Shared Library
5. WHEN 本地同时开发多个 NF module, the Aggregator Repo SHALL 提供 `go.work` 将各 NF Project 纳入同一 Go Workspace

### Requirement 5: 聚合仓职责收敛

**User Story:** AS 集成测试人员, I want 保留一个不含 NF 业务实现的聚合仓, so that 我仍能一键拉起多 NF 做端到端联调

#### Acceptance Criteria

1. WHEN 拆分完成, the Aggregator Repo SHALL 在根目录仅保留编排脚本、配置样例、集成测试、文档、`go.work` 与 `nfs/<nf>/` 目录引用，根目录不存放任一 NF 的业务实现
2. WHEN 集成测试需要某一 NF, the Aggregator Repo SHALL 通过独立构建产物启动该 NF Process，使测试进程与各 NF Process 分属不同操作系统进程
3. WHEN 运维人员在 Aggregator Repo 中执行一键启动, the System SHALL 按依赖顺序分别启动 NRF 及其他已配置的 NF Process
4. IF Aggregator Repo 根目录仍存在某一 NF 的业务实现源码, the System SHALL 将该源码移入对应 `nfs/<nf>/` 目录

### Requirement 6: 接口与兼容性

**User Story:** AS 核心网联调人员, I want 拆分后的 NF 继续使用 3GPP 接口与 free5GC 既有配置语义, so that 现有 RAN/UE 模拟器与配置文件可继续用于联调

#### Acceptance Criteria

1. WHEN Control Plane NF 之间通信, the System SHALL 使用 HTTP/2 SBI，服务发现以 NRF 注册与查询为准
2. WHEN SMF 与 UPF 建立会话控制, the System SHALL 使用 N4 上的 PFCP
3. WHEN AMF 与 RAN 交互, the System SHALL 使用 N2 上的 NGAP/SCTP
4. WHEN UPF 转发用户面数据, the System SHALL 使用 N3 上的 GTP-U 以及面向数据网的 N6
5. WHEN 某一 NF 读取配置, the NF Process SHALL 继续支持该 NF 在 free5GC 中已有的 YAML 配置项语义（SBI 地址、NRF URI、PLMN、S-NSSAI、DNN 等本 NF 相关项）
6. WHEN 拆分完成, 每个 NF Project SHALL 能在其他 NF 已按 3GPP 接口可达的前提下独立完成该 NF 职责范围内的信令或转发处理
