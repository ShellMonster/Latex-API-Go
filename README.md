# MathSVG Go 服务端

本项目是数学公式渲染服务的 Go 端实现，负责 HTTP API、缓存、日志与 Rust 渲染库的对接。服务接收 LaTeX 公式，通过 cgo 调用 Rust 共享库生成 SVG，支持高并发场景。Rust 渲染核心仓库位于 [ShellMonster/Latex-Rust](https://github.com/ShellMonster/Latex-Rust)，需先参考该仓库生成共享库。

## 目录结构
```
Go服务端/
├── cmd/server/main.go            # 入口，装配配置/日志/缓存/HTTP 服务
├── internal/
│   ├── api/                      # HTTP 接口：渲染、健康检查、输入校验
│   ├── cache/                    # BigCache + Redis 缓存封装
│   ├── config/                   # Viper 配置加载与默认值
│   ├── logging/                  # Zap + Lumberjack 日志
│   ├── renderer/                 # 渲染器接口、FFI 实现、占位实现
│   ├── server/                   # Fiber 服务封装与中间件
│   └── pkg/ctxkeys/              # 上下文键定义（请求 ID 等）
├── go.mod / go.sum
├── 性能测试/                     # 压测脚本、报告与基准数据（含 .dylib）
└── render_svg/                   # CentOS 环境编译好的共享库（.so 等）
```

## 快速启动
1. 确保已生成共享库：`cd ../Rust渲染 && ./build.sh`
   - `Go服务端/性能测试/` 下存放 macOS 编译好的 `libformula.dylib`
   - `Go服务端/render_svg/` 下存放 CentOS 编译好的共享库（例如 `libformula.so`）
2. 进入 Go 服务目录：
   ```bash
   cd Go服务端
   go mod tidy
   CGO_ENABLED=1 go run ./cmd/server
   ```
3. 测试接口：
   ```bash
   curl "http://127.0.0.1:8080/render?tex=E%3Dmc%5E2"
   curl "http://127.0.0.1:8080/health"
   ```

## 配置要点
- 配置文件采用 Viper：可通过 `config.yaml` 或环境变量（前缀 `MATHSVG_`）覆盖。
- 主要字段：`server.address`、`server.prefork`、`cache.redis_enabled`、`log.filename` 等。
- Redis 可选，默认 `false`；即使启用失败会自动降级至 BigCache。

## 性能摘要
- 压测与基准详情见 [`性能测试/性能表现报告.md`](性能测试/性能表现报告.md)，涵盖原始数据与测试脚本。
- **HTTP 实测（历史样本）**：200 QPS 下平均延迟约 0.97 ms，P95 1.46 ms，成功率 100%。
- **FFI 基准**（`go test -bench=FFI`）：
  - 简单公式 `E=mc^2`：顺序 917 ns/次，并行 224 ns/次（约 450 万次/秒）。
  - 复杂公式 `\displaystyle \int_{0}^{\infty} ... dx`：顺序 9.07 µs/次，并行 1.87 µs/次（约 53 万次/秒）。
- **Node + KaTeX 对比**（参考 `性能对比报告.md`）：在同硬件上，KaTeX 渲染简/复杂公式需 \~14.1 µs / 245.6 µs，Rust 方案仍快约 15~25 倍，并且输出 SVG 更便于后端缓存与跨平台展示。

> 当前开发机限制重跑 HTTP 压测时监听端口被阻止（`bind: operation not permitted`），建议在目标服务器上按报告附录步骤复测。

## 常见问题
- **找不到共享库**：确认 `Rust渲染/libformula.{so|dylib}` 与 Go 二进制保持相对路径，必要时设置 `LD_LIBRARY_PATH`/`DYLD_LIBRARY_PATH`。
- **cgo 未启用**：编译/运行需 `CGO_ENABLED=1`，否则会回退占位渲染器。
- **端口被占用/权限不足**：Prefork 会 fork 多个进程，端口占用时需释放旧进程或在配置中调整 `server.address`。
