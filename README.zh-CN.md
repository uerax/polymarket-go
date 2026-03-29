# polymarket-go

[English](./README.md) | 中文

一个面向 Polymarket Gamma 与 CLOB API 的轻量 Go 客户端和 CLI。

## 功能

- 从 Gamma API 拉取市场列表
- 在本地按查询词搜索/过滤市场
- 通过 ID 获取单个市场
- 通过 slug 获取单个市场
- 从 CLOB API 获取 token price
- 从 CLOB API 获取 order book

## 环境要求

- Go 1.23.5 或更高版本

## 安装

### 作为库使用

```bash
go get github.com/uerax/polymarket-go/pkg/polymarket
```

### 构建 CLI

```bash
go build -o polymarket-go ./cmd
```

## 库调用示例

```go
package main

import (
    "fmt"
    "log"
    "time"

    polymarket "github.com/uerax/polymarket-go/pkg/polymarket"
)

func main() {
    client := polymarket.NewClient(polymarket.Config{
        Timeout:      10 * time.Second,
        DefaultLimit: 10,
    })

    active := true
    markets, err := client.ListMarkets(polymarket.ListMarketsOptions{
        Limit:  5,
        Active: &active,
        Query:  "bitcoin",
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, market := range markets {
        fmt.Println(market.ID, market.Question)
    }
}
```

## CLI 用法

先构建 CLI：

```bash
go build -o polymarket-go ./cmd
```

### 查询市场列表

```bash
./polymarket-go markets --limit 5 --query bitcoin --active
```

### 通过 ID 获取市场

```bash
./polymarket-go market --id 12345
```

### 通过 slug 获取市场

```bash
./polymarket-go market --slug will-btc-hit-200k
```

### 获取 token price

```bash
./polymarket-go price --token-id <TOKEN_ID> --side buy
```

### 获取 order book

```bash
./polymarket-go book --token-id <TOKEN_ID>
```

## 开发

运行测试：

```bash
go test ./...
```

构建全部包：

```bash
go build ./...
```

不生成二进制直接运行 CLI：

```bash
go run ./cmd markets --limit 1
```

## 项目结构

```text
.
├── cmd/
│   └── main.go
├── pkg/
│   └── polymarket/
│       ├── client.go
│       ├── client_test.go
│       └── model.go
├── go.mod
├── README.md
└── README.zh-CN.md
```

## 说明

- 默认 Gamma Base URL：`https://gamma-api.polymarket.com`
- 默认 CLOB Base URL：`https://clob.polymarket.com`
- CLI 会将结果以格式化 JSON 输出到标准输出。
