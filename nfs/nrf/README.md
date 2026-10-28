# NRF

独立 Network Repository Function 进程。提供 Nnrf_NFManagement 与 Nnrf_NFDiscovery，NF Profile 保存在进程内存中，不依赖 MongoDB 或其他 NF 进程。

```bash
# 构建
make nrf

# 单独运行
./bin/nrf -c config/nrfcfg.yaml
```

默认监听 `http://127.0.0.1:8000`。
