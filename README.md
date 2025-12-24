# natools4go

```bash
go get -u github.com/natholdallas/natools4go
```

## English

| Directory       | Description           | Presumed Usage                                                          |
| :-------------- | :-------------------- | :---------------------------------------------------------------------- |
| **arrs**        | Array & Slice Utils   | Helper functions for slice operations like unique, merge, or filter.    |
| **concur**      | Concurrency Control   | Wraps Goroutine pools, Context control, or thread-safe patterns.        |
| **constraints** | Type Constraints      | Defines interface constraints for Go generics (Type parameters).        |
| **fibers**      | Fiber Framework Logic | Fiber-based logic including JWT, Cache, and Request Header handling.    |
| **fmts**        | Formatting Utils      | Standardizes string formatting, log output, or numerical formats.       |
| **gorms**       | GORM Wrapper          | Handles GORM initialization, transactions, and common CRUD logic.       |
| **jsons**       | JSON Processing       | Utilities for efficient JSON parsing, serialization, or dynamic maps.   |
| **maths**       | Mathematics           | Provides algorithm implementations, high-precision math, or statistics. |
| **rands**       | Randomization         | Generates random strings, OTPs, UUIDs, or cryptographic salts.          |
| **spew**        | Deep Debugging        | Wraps the `go-spew` library for visualizing complex structs in dev.     |
| **strs**        | String Utilities      | Tools for case conversion, naming conventions, and regex matching.      |
| **va**          | Validation            | Short for Validation; used for verifying struct fields or API inputs.   |
| **vipers**      | Config Management     | Wraps the Viper library for managing YAML/JSON configs and envs.        |

## Chinese

| 文件夹名        | 功能描述           | 核心用途推测                                                  |
| :-------------- | :----------------- | :------------------------------------------------------------ |
| **arrs**        | 数组与切片工具     | 提供对切片的去重、合并、过滤等辅助操作。                      |
| **concur**      | 并发控制           | 封装协程池、Context 控制或并发安全的数据处理。                |
| **constraints** | 类型约束           | 定义 Go 泛型所需的 Interface 约束。                           |
| **fibers**      | Fiber Web 框架扩展 | 包含基于 Fiber 的 JWT 鉴权、缓存、Header 处理等中间件或逻辑。 |
| **fmts**        | 格式化工具         | 统一定义字符串格式化、日志输出格式或金额/数字格式。           |
| **gorms**       | GORM 数据库封装    | 封装 GORM 初始化、事务处理及通用 CRUD 逻辑。                  |
| **jsons**       | JSON 处理          | 封装 JSON 的高效解析、动态结构处理或序列化工具。              |
| **maths**       | 数学计算           | 提供常用的算法实现、高精度计算或统计函数。                    |
| **rands**       | 随机数生成         | 生成随机字符串、验证码、UUID 或加密盐。                       |
| **spew**        | 深度调试打印       | 封装 `go-spew` 库，用于在开发环境可视化输出复杂结构体。       |
| **strs**        | 字符串操作         | 包含大小写转换、驼峰命名转换、正则匹配等工具。                |
| **va**          | 数据校验           | 缩写自 Validation，用于验证结构体字段或 API 参数。            |
| **vipers**      | 配置管理           | 封装 Viper 库，用于读取 YAML/JSON 配置文件及环境变量。        |

## todo

- [ ] spew 添加更高效率，可观的 `struct` 输出格式
- [ ] arrs 添加更多函数
- [ ] maths 包拓展
- [x] strs 重构
- [ ] 优化性能，添加注释
