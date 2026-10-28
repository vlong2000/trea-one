# NFProfile

NRF 中登记的网元描述，键为 `nfInstanceId`。

至少包含：
- `nfInstanceId`
- `nfType`（如 AMF、SMF）
- `nfStatus`（如 REGISTERED）

可选 `allowedNfTypes`：发现时按请求方 NF 类型过滤。未设置则允许任意 requester。

数据只存在于当前 NRF 进程内存，进程退出后丢失。
