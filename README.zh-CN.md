# polymarket-go (CLOB SDK)

Polymarket CLOB API 的 Go SDK，基于 TypeScript `clob-client` 语义迁移。

## 特性

- **易用的 API** - 行情数据、交易、账户管理
- **多层认证** - 支持 L1（签名器）和 L2（API 密钥）
- **游标分页** - 适应大结果集
- **自动重试** - 处理短暂故障，与 TS SDK 对齐
- **请求验证** - 完整的请求校验和错误处理

## 环境要求

- Go 1.23.5+

## 安装

```bash
go get github.com/uerax/polymarket-go/polymarket
```

## 快速开始

### 1. 公开行情数据（无需认证）

```go
package main

import (
	"context"
	"fmt"

	"github.com/uerax/polymarket-go/polymarket"
)

func main() {
	ctx := context.Background()

	// 创建客户端（公开数据无需认证）
	client := polymarket.NewClient(
		"https://clob.polymarket.com",
		polymarket.ChainPolygon,
		nil,  // 签名器
		nil,  // API 密钥凭证
	)

	// 获取服务器时间
	ts, err := client.GetServerTime(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("服务器时间: %d\n", ts)

	// 获取代币的订单簿
	book, err := client.GetOrderBook(ctx, "12345")
	if err != nil {
		panic(err)
	}
	fmt.Printf("买单: %d, 卖单: %d\n", len(book.Bids), len(book.Asks))

	// 获取市场价格
	prices, err := client.GetPrice(ctx, "12345", "BUY")
	if err != nil {
		panic(err)
	}
	fmt.Printf("价格: %v\n", prices)
}
```

### 2. 创建 API 密钥（仅需 L1 认证）

```go
// 需要一个 ClobSigner 实现
client := polymarket.NewClient(
	"https://clob.polymarket.com",
	polymarket.ChainPolygon,
	signer, // 你的签名器实现
	nil,
)

creds, err := client.CreateAPIKey(ctx, nil)
if err != nil {
	panic(err)
}
fmt.Printf("API 密钥: %s\n", creds.Key)
```

### 3. 查询账户与订单（L2 认证）

```go
client := polymarket.NewClient(
	"https://clob.polymarket.com",
	polymarket.ChainPolygon,
	signer,  // L1 认证签名
	&polymarket.ApiKeyCreds{
		Key:        "你的-api-密钥",
		Secret:     "你的-secret",
		Passphrase: "你的-passphrase",
	},
)

// 获取开放订单
orders, err := client.GetOpenOrders(ctx, nil, false, "")
if err != nil {
	panic(err)
}
for _, order := range orders {
	fmt.Printf("订单: %s, 方向: %s, 数量: %s\n", order.ID, order.Side, order.Size)
}

// 获取交易历史
trades, err := client.GetTrades(ctx, &polymarket.TradeParams{}, false, "")
if err != nil {
	panic(err)
}
for _, trade := range trades {
	fmt.Printf("交易: %s, 价格: %s\n", trade.ID, trade.Price)
}
```

### 4. 下单（需要 L2 认证）

```go
// 首先用签名器构建和签名订单
signedOrder, err := client.CreateOrder(ctx, polymarket.UserOrder{
	TokenID:    "12345",
	Price:      0.5,
	Size:       100.0,
	Side:       polymarket.SideBUY,
	Expiration: nil,
}, nil)
if err != nil {
	panic(err)
}

// 提交订单
result, err := client.PostOrder(
	ctx,
	signedOrder,
	polymarket.OrderTypeGTC,
	false, // deferExec
	nil,   // postOnly
)
if err != nil {
	panic(err)
}
fmt.Printf("订单已提交: %v\n", result)

// 取消订单
cancelResult, err := client.CancelOrder(ctx, polymarket.OrderPayload{
	OrderID: "订单哈希",
})
if err != nil {
	panic(err)
}
fmt.Printf("取消结果: %v\n", cancelResult)
```

## 客户端初始化

### 基本选项

```go
client := polymarket.NewClient(
	host,    // API 主机 URL (如 "https://clob.polymarket.com")
	chain,   // 链 ID (ChainPolygon=137 或 ChainAmoy=80002)
	signer,  // 签名器 (无需 L1 认证时为 nil)
	creds,   // API 密钥凭证 (无需 L2 认证时为 nil)
	// ... 其他可选客户端选项
)
```

### 可选客户端选项

```go
client := polymarket.NewClient(
	host, chain, signer, creds,
	polymarket.WithHTTPClient(customHTTPClient),
	polymarket.WithThrowOnError(true),           // 抛出 API 错误
	polymarket.WithRetryOnError(true),           // 重试短暂错误
	polymarket.WithGeoBlockToken("你的-token"),  // 地理位置限制绕过
	polymarket.WithUseServerTime(true),          // 使用服务器时间
	polymarket.WithTickSizeTTL(5 * time.Minute), // 行情缓存 TTL
)
```

## API 参考

### 行情数据（公开 - 无需认证）

行情发现和定价端点：

- **GetOrderBook(ctx, tokenID)** - 获取代币订单簿
- **GetOrderBooks(ctx, params)** - 获取多个订单簿
- **GetPrice(ctx, tokenID, side)** - 获取代币的价格
- **GetPrices(ctx, params)** - 获取多个代币的价格
- **GetSpread(ctx, tokenID)** - 获取点差
- **GetMidpoint(ctx, tokenID)** - 获取中间价
- **GetTickSize(ctx, tokenID)** - 获取最小 tick 大小
- **GetNegRisk(ctx, tokenID)** - 检查是否为负风险代币
- **GetFeeRateBps(ctx, tokenID)** - 获取费率（BPS）
- **GetPricesHistory(ctx, params)** - 获取历史价格
- **GetMarkets(ctx, nextCursor)** - 分页市场列表
- **GetMarket(ctx, conditionID)** - 获取市场详情

### 认证管理（L1 认证 - 需要签名器）

API 密钥和认证管理：

- **CreateAPIKey(ctx, nonce)** - 创建新 API 密钥
- **DeriveAPIKey(ctx, nonce)** - 导出现有 API 密钥
- **CreateOrDeriveAPIKey(ctx, nonce)** - 创建或导出（优先创建）
- **CreateReadonlyAPIKey(ctx)** - 创建只读密钥
- **GetReadonlyAPIKeys(ctx)** - 列出只读密钥
- **DeleteReadonlyAPIKey(ctx, key)** - 删除只读密钥
- **ValidateReadonlyAPIKey(ctx, address, key)** - 验证密钥

### 账户与订单（L2 认证 - 需要 API 密钥）

订单和账户信息：

- **GetOrder(ctx, orderID)** - 获取单个订单详情
- **GetOpenOrders(ctx, params, onlyFirstPage, nextCursor)** - 列出开放订单
- **GetTrades(ctx, params, onlyFirstPage, nextCursor)** - 列出交易
- **GetTradesPaginated(ctx, params, nextCursor)** - 单页交易
- **GetAPIKeys(ctx)** - 列出所有 API 密钥
- **GetClosedOnlyMode(ctx)** - 检查账户是否仅平仓
- **GetBalanceAllowance(ctx, params)** - 获取余额和额度

### 订单管理（L2 认证）

下单和管理订单：

- **PostOrder(ctx, order, orderType, deferExec, postOnly)** - 提交单个订单
- **PostOrders(ctx, args, deferExec, defaultPostOnly)** - 批量下单
- **CancelOrder(ctx, payload)** - 按订单 ID 取消
- **CancelOrders(ctx, orderHashes)** - 按哈希批量取消
- **CancelAll(ctx)** - 取消所有订单
- **CancelMarketOrders(ctx, params)** - 按市场/资产取消
- **IsOrderScoring(ctx, params)** - 检查订单是否评分
- **AreOrdersScoring(ctx, params)** - 检查多个订单

### 奖励与收益（L2 认证）

流动性奖励和收益：

- **GetEarningsForUserForDay(ctx, date)** - 日收益明细
- **GetTotalEarningsForUserForDay(ctx, date)** - 日总收益
- **GetRewardPercentages(ctx)** - 奖励配置
- **GetUserEarningsAndMarketsConfig(ctx, ...)** - 详细收益与市场配置
- **GetCurrentRewards(ctx)** - 当前奖励池
- **GetRawRewardsForMarket(ctx, conditionID)** - 特定市场奖励

### RFQ（报价请求 - L2 认证）

RFQ 工作流：

- **CreateRfqRequest(ctx, order, options)** - 创建报价请求
- **CancelRfqRequest(ctx, params)** - 取消请求
- **GetRfqRequests(ctx, params)** - 列出请求
- **CreateRfqQuote(ctx, quote, options)** - 为请求创建报价
- **CancelRfqQuote(ctx, params)** - 取消报价
- **GetRfqRequesterQuotes(ctx, params)** - 列出报价（请求方）
- **GetRfqQuoterQuotes(ctx, params)** - 列出报价（报价方）
- **GetRfqBestQuote(ctx, params)** - 获取最佳报价
- **AcceptRfqQuote(ctx, payload)** - 接受报价
- **ApproveRfqOrder(ctx, payload)** - 批准 RFQ 订单
- **RFQConfig(ctx)** - 获取 RFQ 配置

### Builder API（可选 - 需要 Builder 认证头）

- **CreateBuilderAPIKey(ctx)** - 创建 builder 密钥
- **GetBuilderAPIKeys(ctx)** - 列出 builder 密钥
- **RevokeBuilderAPIKey(ctx)** - 撤销 builder 密钥
- **GetBuilderTrades(ctx, params, nextCursor)** - 获取 builder 交易

## 常见模式

### 分页

许多列表端点支持基于游标的分页：

```go
// 迭代所有结果
var allResults []YourType
nextCursor := ""

for nextCursor != polymarket.EndCursor {
	result, err := client.SomeListEndpoint(ctx, params, false, nextCursor)
	if err != nil {
		panic(err)
	}
	allResults = append(allResults, result...)
	nextCursor = result.NextCursor // 或从响应中设置
}
```

或获取单页：

```go
page, err := client.GetTradesPaginated(ctx, params, "")
if err != nil {
	panic(err)
}
fmt.Printf("交易: %d, 下一页: %s\n", len(page.Trades), page.NextCursor)
```

### 错误处理

```go
result, err := client.PostOrder(ctx, order, orderType, false, nil)
if err != nil {
	// 检查是否是 ApiError
	if apiErr, ok := err.(*polymarket.ApiError); ok {
		fmt.Printf("API 错误: 状态 %d, 正文: %s\n", apiErr.Status, string(apiErr.Body))
	} else {
		fmt.Printf("错误: %v\n", err)
	}
}
```

### 订单类型

```go
const (
	OrderTypeGTC OrderType = "GTC"  // 取消前有效
	OrderTypeGTD OrderType = "GTD"  // 有效期至日期
	OrderTypeFOK OrderType = "FOK"  // 全部成交或取消
	OrderTypeFAK OrderType = "FAK"  // 自动成交并取消
)
```

### 认证级别

| 端点 | 所需认证 | 示例 |
|------|--------|------|
| 行情数据 | 无 | `GetOrderBook`, `GetPrice` |
| 创建 API 密钥 | L1 (签名器) | `CreateAPIKey` |
| 账户信息 | L2 (API 密钥) | `GetOpenOrders`, `GetTrades` |
| 下单 | L2 (API 密钥) | `PostOrder` |
| Builder API | Builder 头 | `GetBuilderTrades` |

## 实现 ClobSigner

`ClobSigner` 接口处理签名生成：

```go
type ClobSigner interface {
	// 获取用户钱包地址
	Address(ctx context.Context) (string, error)

	// 签名 CLOB 认证头（用于 L1 认证）
	SignClobAuth(ctx context.Context, chainID Chain, timestamp int64, nonce int64) (string, error)

	// 签名订单类型数据（用于订单签名）
	SignOrderTypedData(ctx context.Context, payload OrderTypedDataPayload) (string, error)
}
```

## 测试

运行测试：

```bash
# 所有测试
go test ./...

# 仅包测试
go test ./polymarket -v

# 单个测试
go test ./polymarket -run TestThrowOnErrorReturnsApiError -v
```

## 包结构

```text
polymarket/
├── client.go            # 主客户端和 API 方法
├── constants.go         # 端点和游标常量
├── types.go             # 请求/响应模型
├── errors.go            # ApiError 和认证错误
├── http_helpers.go      # HTTP + 错误映射 + 重试 + throwOnError
├── headers.go           # L1/L2 头部生成
├── signer.go            # 签名器抽象
├── order_types.go       # 签名/builder 相关接口
└── http_helpers_test.go # 关键行为对齐测试
```

## 与 TS SDK 对齐

本 SDK 在以下关键方面与 TS `clob-client` 对齐：

- 游标分页（`MA==` / `LTE=`）
- 错误对象映射 + 可选 `throwOnError`
- L1/L2 头部流程
- 查询参数序列化细节（含 RFQ 重复参数风格）

后续完整对齐工作只在 `polymarket` 内继续推进。

## 许可

MIT
