# Implementation Task List: split-free5gc-nfs

## Tasks

- [x] 1. 将 NRF 拆为独立 Go module 与独立进程
  - `nfs/nrf` 独立 `go.mod` / `cmd` / 配置
  - 构建产物 `bin/nrf`，默认 `config/nrfcfg.yaml`
  - 内存存储 NF Profile，提供 Nnrf_NFManagement 与 Nnrf_NFDiscovery
  - 不依赖其他 NF 业务包与 MongoDB
- [ ] 2. 拆分 UDR
- [ ] 3. 拆分 UDM
- [ ] 4. 拆分 AUSF
- [ ] 5. 拆分 NSSF
- [ ] 6. 拆分 PCF
- [ ] 7. 拆分 AMF
- [ ] 8. 拆分 SMF
- [ ] 9. 拆分 UPF
- [ ] 10. 聚合仓一键编排与端到端联调
