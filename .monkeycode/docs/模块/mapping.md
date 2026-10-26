# mapping

包路径：`mml-server/internal/mapping`

`Load(path)` 读取映射文件并构建 `Table`。`Lookup(request)` 精确匹配，未命中返回 `ERROR: unknown command`。

测试：`internal/mapping/mapping_test.go`。
