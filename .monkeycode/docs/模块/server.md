# server

包路径：`mml-server/internal/server`

`New(addr, mapper)` 创建服务。`ListenAndServe(ctx)` 绑定地址并 Accept；`ctx` 取消后关闭 listener，等待 Session 最多 3 秒。`Addr()` 在监听就绪后返回实际地址（测试用随机端口）。

测试：`internal/server/server_test.go`。
