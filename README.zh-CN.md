# polymarket-go (CLOB SDK)

这是一个按 `clob-client` 语义迁移的 Polymarket CLOB Go SDK。

## 范围

本仓库现在只保留一个核心包：

- `polymarket`：CLOB 客户端 SDK（公共接口 + 鉴权 + 订单/奖励/builder/RFQ 相关方法）

旧的 `pkg/polymarket` 与旧 CLI 入口已移除。

## 环境要求

- Go 1.23.5+

## 作为库安装

```bash
go get github.com/uerax/polymarket-go/polymarket
```

## 运行测试

```bash
go test ./...
```

## 包结构

```text
polymarket/
├── client.go            # 主客户端与 API 方法
├── constants.go         # endpoint 与 cursor 常量
├── types.go             # 请求/响应模型
├── errors.go            # ApiError 与鉴权错误
├── http_helpers.go      # HTTP + 错误映射 + 重试 + throwOnError
├── headers.go           # L1/L2 头部构造
├── signer.go            # 签名器抽象
├── order_types.go       # Signature/builder 相关接口
└── http_helpers_test.go # 关键行为对齐测试
```

## 与 TS 对齐说明

当前已对齐以下关键语义：

- cursor 分页（`MA==` / `LTE=`）
- 错误对象映射 + 可选 `throwOnError`
- L1/L2 header 流程
- query 参数序列化细节（含 RFQ 重复参数风格）

后续全量对齐工作只在 `polymarket` 内继续推进。
