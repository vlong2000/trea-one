# Mapping Table

Mapping Table 将 Request Line 精确映射到 Response Line。内存结构为 `map[string]string`，启动后只读。

加载规则：
- 空行与 `#` 开头行忽略
- 合法行：`request<TAB>response`
- `request` 非空；`response` 允许为空
- 无 TAB 或 TAB 左侧为空：跳过并记日志
- 重复请求字符串：后出现的项覆盖先出现的项
- 文件缺失或零条合法映射：启动失败

查询：整行完全相等。未命中返回 `ERROR: unknown command`。
