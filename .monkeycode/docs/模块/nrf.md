# 模块: nrf

独立 Network Repository Function。Go module 路径 `github.com/vlong2000/trea-one/nfs/nrf`，构建产物 `bin/nrf`。

**位置**: `nfs/nrf/`

**职责**:
- 读取 `config/nrfcfg.yaml`
- 在 SBI 地址上提供 Nnrf_NFManagement 与 Nnrf_NFDiscovery
- 将 NF Profile 保存在进程内存

**关键文件**:
- `cmd/main.go` - 进程入口
- `pkg/factory/config.go` - YAML 加载与校验
- `pkg/service/service.go` - 生命周期与信号处理
- `internal/sbi/server.go` - HTTP 路由
- `internal/store/store.go` - 内存 Profile 表

**依赖**: `gopkg.in/yaml.v3`。不依赖其他 NF 业务包。

启动失败（配置缺失、端口占用）以非零码退出。SIGINT/SIGTERM 停止本进程。
