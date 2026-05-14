# Examples

可运行的最小示例集。每个子目录下 `main.go` 单独可运行。

```bash
cd examples/string && go run main.go
cd examples/hash   && go run main.go
cd examples/delay  && go run main.go
cd examples/locker && go run main.go
```

所有示例假设本地 Redis 监听在 `127.0.0.1:6379`，可通过环境变量 `REDIS_ADDR` 覆盖。

| 示例 | 演示内容 |
|------|----------|
| `string/` | 基础 GET/SET、过期时间、Pipeline 批量写入 |
| `hash/`   | HSET / HGetAll / HMSet / HExpire（Redis 7.4+） |
| `delay/`  | DelayQueue 投递、消费、重试、死信 |
| `locker/` | Locker 互斥锁、TryLock、续约 |

## 进阶示例

`examples/` 之外仓库根的 `*_test.go` 与 `delay_test.go` 提供了更复杂的测试驱动用法，适合做正式集成参考。
