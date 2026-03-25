# Go 工具链全量修炼 · 加密货币撮合引擎实战路线图

> **目标**：以一个真实撮合交易引擎为载体，系统掌握 Go 标准库的每一个工具，建立从"会用"到"用好"的完整能力闭环。
>
> **前提**：已安装 Go 1.21+，已安装 Git，了解基本的 HTTP / WebSocket 概念，对交易所订单簿有基础了解（不了解也没关系，本文档会逐步讲解）。
>
> **每个模块均为独立练习题，先独立思考，再查阅提示，最后参考答案。**

---

## 项目概览

### 我们要做什么

构建一个 **MatchEngine** —— 极简加密货币撮合交易引擎（类 Binance / OKX），后端用 Gin 框架，支持：

- **撮合引擎**：限价单 / 市价单的价格-时间优先撮合（核心）
- **订单簿**：实时订单簿（买盘/卖盘）管理
- **用户账户**：余额冻结、持仓管理、流水记录
- **市场数据**：Ticker、K线、深度、成交历史
- **WebSocket**：实时订单簿推送、成交推送、用户委托推送
- **风控**：订单频率限制、持仓限额、余额校验

在这个过程中，**每个功能模块都刻意使用不同的标准库**，让你在实战中自然掌握工具链。

### 章节速查

| 章节 | 主题 | 核心库 |
|------|------|--------|
| 第一章 | 项目初始化与撮合引擎架构 | `strconv`, `strings` |
| 第二章 | 订单簿与深度数据 | `bytes`, `bufio` |
| 第三章 | Ticker 符号与 Unicode | `unicode`, `utf8` |
| 第四章 | 订单反射验证 | `reflect` |
| 第五章 | 热点路径零拷贝 | `unsafe` |
| 第六章 | 撮合逻辑测试 | `testing` |
| 第七章 | 错误处理与风控 | `errors`, `fmt` |
| 第八章 | 上下文取消与超时 | `context` |
| 第九章 | HTTP 客户端与外部数据源 | `net/http`, `io` |
| 第十章 | Gin 框架核心与中间件 | `gin` |
| 第十一章 | 持久化：MySQL + Redis | `database/sql` |
| 第十二章 | GitHub 协作与 CI/CD | `git`, `gh` |
| 第十三章 | 综合实战：构建完整撮合系统 | 全部库综合 |

---

## 第一章 · strconv + strings：订单解析与符号处理

### 背景

撮合引擎接收的所有外部数据（HTTP 请求、WebSocket 帧、TCP 二进制协议）最初都是字符串。`strconv` 将字符串转为数值（价格/数量），`strings` 处理符号验证、过滤、分隔。这两个库是整个系统的输入第一关。

### 核心 API

```go
// strconv
strconv.Atoi(s string) (int, error)
strconv.ParseFloat(s string, bitSize int) (float64, error)
strconv.FormatFloat(f float64, fmt byte, prec, bitSize int) string
strconv.ParseInt / FormatInt / ParseUint / FormatUint

// strings
strings.TrimSpace / strings.Trim
strings.Split / strings.SplitN / strings.Join
strings.ToUpper / strings.ToLower
strings.Contains / strings.HasPrefix / strings.HasSuffix
strings.Replace / strings.ReplaceAll
strings.Builder  // 高效拼接
```

### 练习题

---

#### 1-1：订单价格与数量解析

**难度**：⭐

**题目**：实现 `ParseOrderParams(c *gin.Context) (*OrderParams, error)`，从 HTTP 请求中解析下单参数：

```go
type OrderParams struct {
    Symbol    string  // 交易对，如 "BTCUSDT"
    Side      string  // "BUY" 或 "SELL"
    Type      string  // "LIMIT" 或 "MARKET"
    Price     float64 // 限价单价格（精确到 0.01）
    Quantity  float64 // 数量（精确到 0.0001）
    ClientOID string  // 客户端订单ID（可选，最大64字符）
}
```

**要求**：
- 使用 `strconv.ParseFloat` 解析 `Price` 和 `Quantity`，设置 `bitSize: 64`
- 处理解析错误：`ParseFloat("abc")` 返回 `NaN`，需用 `math.IsNaN()` 判断
- 精度校验：`Price > 0`，`Quantity > 0`，`Price` 最多 2 位小数，`Quantity` 最多 4 位小数
- `Symbol` 必须为标准格式：`BASE*QUOTE`（如 BTC/USDT 交易对格式化为 `BTCUSDT`），用 `strings.ToUpper` 规范化
- `Side` 只接受 `"BUY"` / `"SELL"`（大小写不敏感）
- 编写 10+ 个表驱动单元测试（正常单、非法价格、非法数量、空 Symbol、精度超限等）

**扩展**：
- 用 `fmt.Sprintf("%.2f", price)` 将价格规范化到固定精度

---

#### 1-2：交易对符号解析与校验

**难度**：⭐⭐

**题目**：实现一个交易对管理器 `SymbolValidator`：

```go
type SymbolConfig struct {
    BaseAsset    string  // 基础资产，如 "BTC"
    QuoteAsset   string  // 计价资产，如 "USDT"
    PricePrec    int     // 价格精度（小数位数）
    QtyPrec      int     // 数量精度
    MinQty       float64 // 最小下单量
    MaxQty       float64 // 最大下单量
    MinNotional  float64 // 最小成交额（price * qty）
}
```

**要求**：
- 函数 `ParseSymbol(symbol string) (base, quote string, err error)`：
  - 支持 `BTCUSDT`（无分隔符）、`BTC/USDT`（斜杠分隔）、`BTC-USDT`（横杠分隔）
  - 用 `strings.ToUpper` 规范化，`strings.TrimSpace` 去空格
  - 用 `strings.Index` 找到分隔符位置，`strings.SplitN` 切分
  - `QuoteAsset` 必须在预设的白名单内：`USDT, BUSD, BTC, ETH`
- 函数 `BuildSymbol(base, quote string) string`：反向拼接为标准格式
- 用 `strings.Builder` 高效拼接日志输出
- 至少 15 个测试用例（正常格式、各种分隔符、空串、大小写混合、不在白名单等）

---

#### 1-3：订单簿深度数据格式化

**难度**：⭐⭐

**题目**：实现 `FormatDepthData(bids, asks [][]float64, limit int) string`，将订单簿深度数据格式化为文本表格输出（用于日志和调试）：

```
Symbol: BTCUSDT | Best Bid: 65000.50 | Best Ask: 65001.00
================= ORDER BOOK =================
   BID PRICE     |   ASK PRICE   |  BID QTY  |  ASK QTY
  65000.50       |  65001.00     |   1.2345  |   0.8765
  64999.00       |  65002.50     |   2.1000  |   1.5000
=============================================
Spread: 0.50 (0.0008%)
```

**要求**：
- 用 `fmt.Sprintf` 格式化价格和数量到指定小数位
- 用 `strings.Repeat` 生成分隔线（`strings.Repeat("=", 20)`）
- 用 `strings.Builder` 按行拼接，避免频繁字符串分配
- 用 `strconv.Itoa` 将整数索引转为字符串
- 按 `limit` 参数截断深度（只显示最优 N 档）
- 计算 Spread（买卖价差）和 Spread 百分比

**扩展**：
- 用 `strings.Fields` 按空格分割格式化后的文本，重新解析为结构化数据（反向验证）
- 生成 CSV 格式的导出（`strings.Join` + 换行符）

---

## 第二章 · bytes 与 bufio：二进制协议与内存管理

### 背景

撮合引擎的撮合核心（matching engine）是性能热点，每次成交都可能涉及数十次内存分配。`bytes` 包操作字节切片，`bufio` 提供缓冲读写，是实现高性能订单处理和协议解析的关键工具。

### 核心 API

```go
// bytes
bytes.Buffer{}  // 可变长缓冲
bytes.NewBuffer / bytes.NewReader
bytes.Join(s [][]byte, sep []byte) []byte
bytes.Split / bytes.SplitN
bytes.Contains / bytes.Equal / bytes.Compare
bytes.Runes(b []byte) []rune
*bytes.Reader.Reset(b []byte)

// bufio
bufio.NewReader / bufio.NewWriter
bufio.NewReaderSize(r io.Reader, size int) *Reader
bufio.Scanner  // 按行/词扫描
```

### 练习题

---

#### 2-1：内存池化（Buffer Pool）

**难度**：⭐⭐

**题目**：实现一个 `OrderBufferPool`，复用 `bytes.Buffer` 减少 GC 压力：

```go
type OrderBufferPool struct {
    pool sync.Pool
    size int  // buffer 初始大小
}
```

**要求**：
- `Get() *bytes.Buffer`：从 Pool 获取，自动 `Reset()`
- `Put(b *bytes.Buffer)`：归还，超大 buffer（> 1MB）不回收
- 在撮合引擎的成交记录写入路径中使用 Pool（每笔成交分配一个临时 buffer）
- 编写 Benchmark：
  - Baseline：每次 `bytes.Buffer{}` 新建
  - Pool 版：使用 `OrderBufferPool`
  - 迭代 5,000,000 次，报告 `ns/op` 和 `allocs/op`
  - 用 `b.ReportMetric` 报告自定义指标

---

#### 2-2：二进制成交记录序列化

**难度**：⭐⭐⭐

**题目**：实现一个自定义二进制协议，将成交记录序列化为 `[]byte`：

**成交记录格式**（固定 48 字节）：
```
[4字节: 成交ID uint32]
[8字节: 价格 uint64, 精度 1e-8]
[8字节: 数量 uint64, 精度 1e-8]
[4字节: 买方用户ID uint32]
[4字节: 卖方用户ID uint32]
[8字节: 时间戳 int64 (Unix毫秒)]
[4字节: 交易对哈希 uint32]
[4字节: 方向 uint8 + 填充]
[4字节: 保留字段]
```

**要求**：
- 实现 `SerializeTrade(t *Trade) ([]byte, error)` 和 `DeserializeTrade(data []byte) (*Trade, error)`
- 用 `bytes.Buffer` 组装字节序列，`binary.Write` / `binary.Read` 写入基本类型
- 用 `bytes.NewReader` + `bufio.NewReader` 从反序列化流中读取
- 用 `bytes.Equal` 比较两笔成交是否完全一致
- 用 `bytes.SplitN` 解析批次的二进制记录
- 边界测试：数据不足 48 字节、数据被截断、字节序错误

---

#### 2-3：CSV/K线数据解析

**难度**：⭐⭐

**题目**：用 `bufio.Scanner` 实现 K 线历史数据的流式解析：

**要求**：
- 读取 CSV 文件（格式：`timestamp,open,high,low,close,volume`），每行一条 K 线
- 用 `bufio.NewScanner` + `scanner.Split(bufio.ScanLines)` 按行扫描
- 用 `strings.Split` 分割逗号，`strconv.ParseFloat` 解析每个字段
- 用 `bufio.Reader` 包装 `*os.File`，用 `os.Open` 打开文件
- 处理文件过大（GB 级别）：不允许一次性读取全文件，必须流式处理
- 用 `sync.Pool` 复用 `bufio.Reader`，减少 goroutine 栈分配

**扩展**：
- 实现 `FlushToDB()` 批量写入（每 1000 条 commit 一次）

---

## 第三章 · unicode 与 utf8：国际化与符号系统

### 背景

交易对符号（如 `BTCUSDT`、`ETHBTC`）、资产名称、公告内容都涉及 Unicode。Go 的 `string` 是 UTF-8 字节序列，`rune` 是 Unicode 码点。正确处理多语言和 emoji 资产名称是健壮系统的标志。

### 核心 API

```go
// unicode
unicode.IsLetter(r) / unicode.IsDigit(r) / unicode.IsUpper/IsLower/IsSpace/IsPunct
unicode.ToUpper / ToLower / ToTitle
unicode.SimpleFold(r rune) rune  // 遍历同一字符的Unicode变体

// utf8
utf8.RuneCountInString(s) int
utf8.DecodeRuneInString(s) (r rune, size int)
utf8.DecodeRune / EncodeRune
utf8.Valid / ValidString
utf8.FullRune
```

### 练习题

---

#### 3-1：资产名称 Unicode 分类过滤

**难度**：⭐⭐

**题目**：实现 `ValidateAssetName(name string) (valid bool, reason string)`：

**要求**：
- 资产名称只能是字母（Latin/CJK）、数字，不能包含空格、标点、控制字符
- 用 `unicode.IsLetter`、`unicode.IsDigit` 逐 rune 判断
- 用 `utf8.RuneCountInString` 验证字符数量（最长 20 个 Unicode 码点）
- 用 `unicode.SimpleFold` 处理特殊大小写折叠（如土耳其文的 `i` / `İ` 问题）
- 检测 emoji：`unicode.In(r, unicode.M)` 或范围 `r >= 0x1F300 && r <= 0x1F9FF`（早期版本 Go 无此函数，手写判断）
- 至少 15 个测试用例（中英文资产名、带 emoji、带空格、带标点、空串、超长）

**示例**：
```
"BTC"      → valid
"以太坊"   → valid
"USDT🔥"   → invalid (emoji)
"BTC USDT" → invalid (space)
```

---

#### 3-2：交易对符号 Unicode 安全校验

**难度**：⭐⭐

**题目**：扩展练习 1-2，在符号校验中加入 Unicode 边界检查：

**要求**：
- 用 `utf8.ValidString(symbol)` 验证符号是有效的 UTF-8（防止注入攻击）
- 用 `utf8.DecodeRuneInString` 获取第一个 rune 和字节长度，验证符号不是以多字节字符开头
- 用 `unicode.ToUpper` 逐 rune 处理后，再用 `strings.Builder` 拼接（对比直接 `strings.ToUpper` 的差异）
- 构造恶意输入：`BTC\u0000USDT`（含 NULL 字节）、`BTC\u200bUSDT`（零宽空格）
- 检测零宽空格（`\u200B`）、BOM（`\uFEFF`）、反向文本（`\u202E`）等隐藏字符

**扩展**：
- 实现 `NormalizeSymbol(s string) string`：移除所有非 ASCII 字母数字字符，用 `strings.Builder` 重建

---

#### 3-3：UTF-8 编码一致性验证

**难度**：⭐⭐⭐

**题目**：实现 `ValidateUTF8Consistency(data []byte) (bool, []int, error)`：

**要求**：
- 遍历 `[]byte`，用 `utf8.FullRune` 判断当前字节位置是否有完整的 rune
- 用 `utf8.DecodeRune` 尝试解码，捕获 `utf8.ErrRuneTooShort` / `utf8.ErrRuneInvalid`
- 返回：是否有效、所有无效字节的位置列表、第一个错误信息
- 手写验证逻辑（**不用** `utf8.Valid` 和 `strings.ToValidUTF8`）
- 构造测试用例：正常 UTF-8、截断的多字节序列、过长编码（4字节以上）、非法字节值

---

## 第四章 · reflect：动态反射与订单验证

### 背景

`reflect` 是构建通用验证器、序列化工具、DI 容器的基石。撮合引擎需要处理数十种不同结构（限价单、市价单、冰山单、止盈止损单），用反射实现统一的参数验证和结构转换。

### 核心 API

```go
reflect.TypeOf(i interface{}) reflect.Type
reflect.ValueOf(i interface{}) reflect.Value
v.Kind()     // reflect.Struct, reflect.Slice, reflect.Map ...
v.NumField() / v.Field(i) / v.FieldByName
v.Type().Field(i).Tag  // 获取 struct tag
v.Elem()               // 解引用
v.Interface()          // 转回 interface{}
v.CanSet() / v.SetInt() / v.SetString()
v.Len() / v.Index(i) / v.MapKeys
v.Method(i).Call(args) // 动态调用方法
```

### 练习题

---

#### 4-1：通用订单结构体验证器

**难度**：⭐⭐⭐

**题目**：用 `reflect` 实现一个通用订单验证器，支持 struct tag 声明验证规则：

```go
type LimitOrder struct {
    Symbol   string  `validate:"required,symbol,max=20"`
    Price    float64 `validate:"required,gt=0,precision=2"`
    Quantity float64 `validate:"required,gt=0,gt_field=MinQty,precision=4"`
    Side     string  `validate:"required,oneof=BUY SELL"`
    Type     string  `validate:"required,oneof=LIMIT MARKET STOP"`
}
```

**要求**：
- 实现 `ValidateOrder(v interface{}) []FieldError`
- 支持验证规则（用 `v.Type().Field(i).Tag.Get("validate")` 读取）：
  - `required`：字段非零值
  - `gt=X` / `lt=X`：大于/小于阈值
  - `gt_field=X`：大于结构体中另一个字段的值（用于 Quantity > MinQty）
  - `precision=N`：小数精度不超过 N 位（用字符串处理验证）
  - `max=N`：字符串最大长度
  - `oneof=a,b,c`：枚举值
- 用 `strings.Split` 分割标签规则，`strings.TrimSpace` 清理空格
- 用 `strconv.ParseFloat` 解析规则中的数值
- 在 Gin 中间件中集成，作为所有下单接口的统一入口

---

#### 4-2：动态订单类型转换

**难度**：⭐⭐⭐

**题目**：实现 `ConvertOrderToMap(order interface{}) map[string]interface{}`：

**要求**：
- 遍历结构体字段（`v.NumField()` + `v.Field(i)`）
- 用 `v.Type().Field(i).Tag.Get("json")` 获取 JSON 字段名（无 tag 则用字段名）
- 将每个字段值转为 `interface{}`（通过 `v.Field(i).Interface()`）
- 特殊处理：`time.Time` 字段格式化为 RFC3339 字符串
- 特殊处理：`decimal.Decimal` 类型（如果有）转为 string
- 递归处理嵌套结构体（用 `v.Field(i).Kind() == reflect.Struct`）
- 用 `fmt.Sprintf` 将数值格式化统一输出（浮点数处理）

**扩展**：
- 实现反向：`MapToOrder(data map[string]interface{}, orderType interface{}) (interface{}, error)`
- 实现忽略零值字段（`omitempty` tag）：用 `v.Interface()` 判断是否为 Go 零值

---

#### 4-3：撮合事件动态分发

**难度**：⭐⭐⭐⭐

**题目**：实现一个撮合事件分发器（类似简单的事件总线）：

**要求**：
- 定义事件类型：`MatchEvent`、`CancelEvent`、`TradeEvent`、`OrderBookUpdateEvent`
- 用 `reflect.MakeFunc` 为每种事件类型动态创建处理函数
- 用 `reflect.ValueOf(handler).MethodByName(methodName).Call` 动态调用对应的处理方法
- 实现 `Subscribe(eventType reflect.Type, handler interface{})`：
  - 用 `v.Type().NumMethod()` 遍历 handler 的所有方法
  - 用 `v.Type().Method(i).Type` 检查方法签名是否符合事件处理函数类型
  - 自动注册匹配的方法
- 用 `reflect.ValueOf(event).Type()` 获取事件类型，查找对应处理器

---

## 第五章 · unsafe：内存操作的禁区

### 背景

撮合引擎的成交路径是整个系统的性能热点，每笔成交涉及大量订单簿操作和内存分配。`unsafe` 允许零拷贝的类型转换（string ↔ []byte），但必须严格遵守使用规范。这是理解 Go 内存模型的关键。

**⚠️ 警告：生产代码中非必要不使用。以下练习仅用于深入理解内存模型。**

### 核心 API

```go
unsafe.Pointer(p *T) unsafe.Pointer
unsafe.Add(ptr unsafe.Pointer, n uintptr)  // Go 1.17+
unsafe.Offsetof(st.field) uintptr
unsafe.Sizeof(v T) uintptr

// 零拷贝 string ↔ []byte
func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
```

### 练习题

---

#### 5-1：零拷贝字符串转换

**难度**：⭐⭐⭐

**题目**：实现两个零拷贝函数，并进行完整的性能分析：

```go
func StringToBytes(s string) []byte
func BytesToString(b []byte) string
```

**要求**：
- 使用 `unsafe.Pointer` + 指针类型转换实现（见上方 API）
- 写 Benchmark（5,000,000 次迭代）：
  1. Baseline：`[]byte(s)` 和 `string(b)`（会分配内存）
  2. Unsafe 零拷贝版本
- 用 `testing.B.ReportMetric` 报告 `ns/op`、`allocs/op`
- 验证转换正确性：转换前后的内容完全一致（`bytes.Equal` / `==`）
- **警告**：Unsafe 转换后的 `[]byte` 不能修改（会导致数据竞争和内存破坏）

**扩展**：
- 在撮合引擎的日志路径中使用 StringToBytes（高频写日志时减少分配）
- 对比 `bytes.Buffer.String()` 和直接 unsafe 转换的性能差异

---

#### 5-2：结构体内存布局分析

**难度**：⭐⭐⭐⭐

**题目**：实现 `AnalyzeStructLayout(v interface{}) []FieldLayout`：

```go
type FieldLayout struct {
    Name       string
    Type       string
    Offset     uintptr
    Size       uintptr
    Alignment  uintptr
    IsPadded   bool
    TotalSize  uintptr
}
```

**要求**：
- 用 `reflect.TypeOf(v).Elem()` 获取 struct 类型
- 用 `unsafe.Sizeof` 获取整个结构体大小
- 用 `unsafe.Offsetof` 获取每个字段的偏移量
- 判断 padding：当前字段偏移量不是上一个字段大小对齐的整数倍时，存在 padding
- 分析标准库中的关键结构：
  - `time.Time`：64 字节，验证其内存布局（3 个 int64 + 1 个 int32）
  - `sync.Mutex`：8 字节，验证其内部结构（1 个 int32 state + 1 个 int32 sema）
  - 你的 `Order` 结构体：分析每个字段的内存布局和总大小

**扩展**：
- 设计一个 "紧凑版 Order" 结构体，通过调整字段顺序最小化 padding（手工对齐）
- 对比紧凑版 vs 原始版的 `unsafe.Sizeof` 结果

---

#### 5-3：Slice Header 手动构造

**难度**：⭐⭐⭐⭐

**题目**：通过 `unsafe` 手动操作 slice 的内部结构：

```go
type SliceHeader struct {
    Data uintptr  // 指向底层数组
    Len  int      // 长度
    Cap  int      // 容量
}
```

**要求**：
- 将 `[]int{10, 20, 30}` 的地址强转为 `*SliceHeader`
- 读取并验证 Data、Len、Cap 的值
- 手动修改 Len（不调用 `append`），体会 slice "长度 vs 容量" 的边界
- 验证写入超出 Len 但未超出 Cap 的位置（"未初始化写入"）的危险性
- 实现一个 "只读视图" 的 slice：不复制底层数组，共享内存

**扩展**：
- 分析 slice 的扩容机制：当 Cap 不足时，`append` 会申请多大的新数组？（用 Benchmark 观察）

---

## 第六章 · testing：撮合引擎质量保障

### 背景

撮合引擎是状态机，有大量边界条件和并发场景（一人下买单、一人下卖单，同一价格两单先来后到）。没有测试的撮合引擎 = 没有信任。Go 的 `testing` 框架提供单元测试、基准测试、Fuzzing 测试的完整工具链。

### 核心 API

```go
// 基础测试
func TestXxx(t *testing.T)
func BenchmarkXxx(b *testing.B)
func ExampleXxx()  // 示例测试

// 子测试
t.Run(name string, f func(t *testing.T))
t.Parallel()
t.Fatal / t.Fatalf
t.Error / t.Errorf
t.Skip / t.Skipf

// Benchmark
b.ReportAllocs()
b.ResetTimer()
b.RunParallel(func(pb *testing.PB){})

// Fuzzing (Go 1.18+)
func FuzzXxx(f *testing.F)
```

### 练习题

---

#### 6-1：撮合引擎核心逻辑测试

**难度**：⭐⭐

**题目**：为撮合引擎实现完整的单元测试：

```go
type OrderBook struct {
    Bids PriceLevelMap  // map[price]orders ( descending )
    Asks PriceLevelMap
    mu   sync.RWMutex
}

type PriceLevel []*Order
type Order struct {
    ID       int64
    Symbol   string
    Side     string  // "BUY" or "SELL"
    Price    float64
    Quantity float64
    Filled   float64
    Status   string  // "NEW", "PARTIAL", "FILLED", "CANCELED"
    UserID   int64
    Time     int64   // Unix毫秒，用于价格-时间优先级
}
```

**要求**：
- 测试用例（表驱动，至少 15 个）：
  1. 买单价格 >= 卖单价格 → 立即成交（taker 吃单）
  2. 买单价格 < 卖单价格 → 挂单成功，无成交
  3. 数量部分成交（Partial Fill）
  4. 数量完全成交（Full Fill）
  5. 多档卖单 → 依次成交
  6. 价格相同、时间不同 → 时间早的优先成交
  7. 价格不同、时间相同 → 价格更优的优先成交
  8. 撤单：移除未成交部分
  9. 余额不足 → 拒绝下单
  10. 持仓不足（卖单） → 拒绝
  11. 冰山单测试（隐藏数量，仅显示部分）
  12. 市价单测试（以最优对手价成交）
  13. 撮合后订单簿状态正确性
  14. 重复订单ID → 拒绝
  15. 撮合后账户余额变动正确性
- 用 `t.Run` 子测试组织，每个子测试命名清晰
- 用 `sync.WaitGroup` 测试并发撮合（多 goroutine 同时下单）
- 用 `require.Equal` / `assert.Equal` 验证状态（用 `stretchr/testify` 或手写断言）

---

#### 6-2：性能基准测试

**难度**：⭐⭐⭐

**题目**：对撮合引擎进行完整的性能分析：

**要求**：
- `BenchmarkMatchEngine_SingleOrder`：单次撮合 1000 档订单簿的性能
- `BenchmarkMatchEngine_Concurrent`：100 goroutine 并发下单的性能
- 用 `b.ReportAllocs()` 报告内存分配
- 对比不同订单数量（10档、100档、1000档、10000档）下的性能曲线
- 用 `b.RunParallel` 实现真实的并发压力测试
- 用 `go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof` 生成分析文件
- 用 `go tool pprof` 分析 CPU 和内存热点（在命令行交互或生成 SVG 图）

**扩展**：
- 对比 `sync.Mutex` vs `sync.RWMutex` 在撮合引擎中的性能
- 对比 `map[int64]*Order` vs 排序数组（手动实现）作为价格档位的存储

---

#### 6-3：Fuzzing 测试

**难度**：⭐⭐⭐

**题目**：为订单解析器添加 Fuzzing 测试：

**要求**：
- 实现 `FuzzParseOrderParams(f *testing.F)`：
  - 用 `f.Add("BTCUSDT", "BUY", "LIMIT", "65000.50", "1.2345", "")` 提供种子输入
- Fuzzer 自动生成各种畸形输入：随机字节、超长字符串、负数、科学计数法、`NaN`、`Infinity`、特殊字符等
- 用 `f.Fuzz` 替代 `f.Add` 后，Golang 会自动生成随机输入
- 捕获 panic，确保解析函数不会崩溃（用 `defer recover`）
- 对每次生成的输入验证：要么成功解析，要么返回有意义的错误，绝不能 panic

---

#### 6-4：Golden File 测试（成交记录快照）

**难度**：⭐⭐

**题目**：用 Golden File 方式测试成交记录的 JSON 输出：

**要求**：
- 第一次运行测试时，若 `testdata/trade_snapshot.golden` 不存在，自动生成
- 后续运行测试时，对比实际 JSON 输出与 Golden 文件内容
- 用 `os.ReadFile` / `os.WriteFile` 读写 Golden 文件
- 用 `filepath.Join` 构建路径，`filepath.Dir` 获取目录
- 用 `json.MarshalIndent` 格式化输出（两次运行格式一致）
- 每次修改成交记录格式后，只需 `go test -update` 即可更新 Golden 文件

---

## 第七章 · errors 与 fmt：错误处理与风控

### 背景

撮合引擎的错误处理直接关系到资金安全。订单拒绝、余额不足、价格超出限制等错误必须精准分类、快速响应、正确上报。Go 1.13+ 的 `errors.Is` / `errors.As` 配合 `fmt.Errorf` 的 `%w` 包装，是构建分层错误体系的基础。

### 核心 API

```go
errors.New(s string) error
errors.Is(err, target error) bool
errors.As(err error, target interface{}) bool
errors.Join(errs ...error) error  // Go 1.20+
fmt.Errorf("...: %w", err)         // 错误包装
```

### 练习题

---

#### 7-1：分层错误类型体系

**难度**：⭐⭐

**题目**：为撮合引擎设计完整的错误类型体系：

```go
// 顶层接口
type OrderErrorCode interface {
    Code() int         // 错误码
    HTTPStatus() int   // 对应 HTTP 状态码
    Retryable() bool   // 是否可重试
}

// 具体错误类型
type InsufficientBalanceError   { Asset string; Available, Required float64 }
type InvalidPriceError          { Price float64; MinPrice, MaxPrice float64 }
type InvalidQuantityError       { Qty float64; MinQty, MaxQty float64 }
type OrderNotFoundError         { OrderID int64 }
type OrderAlreadyFilledError    { OrderID int64 }
type SymbolNotExistError        { Symbol string }
type MarketClosedError          { Symbol string }
type RateLimitError             { UID int64; Limit int; Window time.Duration }
```

**要求**：
- 每个错误类型实现 `Error() string` 和 `OrderErrorCode` 接口
- 用 `errors.New` 或自定义构造函数创建错误
- 用 `errors.As` 在 API 层按类型处理（将不同错误映射为不同 HTTP 响应）
- 在 Gin 中间件中统一处理：根据 `Code()` 返回 JSON 错误体：
  ```json
  {"code": 1001, "msg": "Insufficient balance", "data": {"available": "10.00", "required": "20.00"}}
  ```
- 错误码规范：
  - `1001-1099`：资金相关错误
  - `2001-2099`：订单参数错误
  - `3001-3099`：市场错误
  - `5001-5099`：系统错误

---

#### 7-2：风控检查中间件

**难度**：⭐⭐⭐

**题目**：实现风控检查中间件 `RiskControlMiddleware`：

**要求**：
- 检查用户下单频率：每 10 秒最多 100 单，超过返回错误码 `5001`
- 检查用户单笔订单最大金额：不超过账户余额的 200%（带杠杆场景）
- 检查总持仓限额：单个交易对持仓不超过总资产的 50%
- 用 `context.WithValue` 传递风控上下文
- 用 `fmt.Sprintf` 格式化错误日志：
  ```
  [RISK] user=123 reject_order reason=balance_too_low available=100.00 required=200.00 symbol=BTCUSDT
  ```
- 用 `strings.Builder` 高效拼接多行日志（每分钟汇总一次）

---

## 第八章 · context：超时取消与链路追踪

### 背景

撮合引擎中，每笔订单都有时效性（TTL）。WebSocket 连接断开时，正在处理的订单应立即取消。数据库查询超时（5s）、外部价格源拉取超时（1s）都需要 `context` 统一管理。`context` 还在分布式追踪中承载 trace_id。

### 核心 API

```go
context.Background() / context.TODO()
context.WithCancel(parent) (ctx, cancel)
context.WithTimeout(parent, d time.Duration) (ctx, cancel)
context.WithDeadline(parent, t time.Time) (ctx, cancel)
context.WithValue(parent, key, val) context.Context
ctx.Err()  // Canceled / DeadlineExceeded
```

### 练习题

---

#### 8-1：订单处理超时控制

**难度**：⭐⭐

**题目**：模拟一个订单撮合流程，涉及多个步骤：

**撮合流程（三个步骤可并发执行）：
- 步骤 A：余额校验（耗时 10ms）
- 步骤 B：风控检查（耗时 5ms）
- 步骤 C：持仓检查（耗时 8ms）

**要求**：
- 总超时时间：20ms
- 任意一个步骤超时则整体失败
- 用 `context.WithTimeout` + `select` + `ctx.Done()` 实现
- 用 `sync.WaitGroup` 等待所有 goroutine 完成
- 记录每个步骤的实际耗时（用 `time.Since(start)`）
- 实现"快速失败"：步骤 A 失败后，步骤 B/C 应立即取消（不需要等待）

**扩展**：
- 实现"所有步骤都成功才继续"的语义（类似 `Promise.all`）
- 实现"任意一个成功即可"的语义（取最快返回的结果）

---

#### 8-2：WebSocket 连接链路追踪

**难度**：⭐⭐

**题目**：在 WebSocket 消息处理中实现 trace_id 追踪：

**要求**：
- Gin 中间件：从请求头 `X-Request-ID` 读取 trace_id，没有则生成（用 `strconv.FormatInt(time.Now().UnixNano(), 36)`）
- 用 `context.WithValue(ctx, traceIDKey, traceID)` 注入 trace_id
- 在日志中输出：`{"trace_id":"xxx","event":"match","duration_ms":5,"orders_matched":2}`
- WebSocket handler 中通过 `c.Request.Context()` 获取 trace_id
- 撮合结果推送时，在 WebSocket 消息中包含 trace_id

**扩展**：
- 用 `context.WithTimeout` 为 WebSocket 连接设置最大生命周期（30 分钟）
- 连接断开时，自动取消所有该连接发起的未完成订单

---

## 第九章 · net/http、io、os：HTTP 客户端与文件 I/O

### 背景

撮合引擎需要从外部数据源（币安 / OKX 公开 API）拉取参考价格、K线数据，进行套利监控或基准测试。`net/http` 是标准 HTTP 客户端，`io`/`os` 是文件操作的基础。

### 核心 API

```go
// net/http 客户端
http.Get(url) (*http.Response, error)
http.NewRequest(method, url, body) (*http.Request, error)
http.Client{Timeout, Transport}
http.DefaultClient

// io
io.ReadAll(r) ([]byte, error)
io.Copy(w, r) (int64, error)
io.ReadFull(r, buf) (int, error)
io.LimitReader(r, n) io.Reader
io.MultiReader(m ...io.Reader) io.Reader
io.TeeReader(r, w) io.Reader

// os
os.Open / os.Create
os.ReadFile / os.WriteFile
os.MkdirAll(path, perm)
os.Stat(name) (FileInfo, error)
```

### 练习题

---

#### 9-1：外部价格源 HTTP 客户端

**难度**：⭐⭐

**题目**：实现一个外部价格源客户端 `PriceSourceClient`：

```go
type PriceSourceClient struct {
    BaseURL   string
    Client    *http.Client
    Symbols   []string  // 订阅的交易对列表
}
```

**要求**：
- 实现 `GetTicker(symbol string) (*Ticker, error)`：
  - 调用外部 API（如 Binance 公开 ticker 接口：`https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT`）
  - 解析 JSON 响应（`json.Unmarshal` 到 `Ticker` 结构体）
  - 设置请求超时：`context.WithTimeout(ctx, 3*time.Second)`
- 实现 `GetKlines(symbol string, interval string, limit int) ([]Kline, error)`：
  - 调用 Binance K线接口
  - 解析嵌套数组响应
  - 用 `io.LimitReader` 限制响应大小（防止攻击）
- 用 `httptest.NewServer` 启动 Mock 服务器，实现完整的集成测试
- 实现自动重试（网络抖动时最多重试 3 次，间隔 100ms/200ms/400ms）

---

#### 9-2：订单簿快照持久化

**难度**：⭐⭐⭐

**题目**：实现订单簿快照的定时落盘（用于故障恢复）：

**要求**：
- 每 5 秒将当前订单簿快照写入文件
- 文件名：`orderbook_{symbol}_{timestamp}.snapshot`
- 用 `os.MkdirAll` 创建快照目录
- 用 `os.Create` + `bufio.NewWriter` 流式写入（不用 `os.WriteFile` 一次性加载）
- 文件格式：自定义文本格式（每行一个档位：`price|bid_qty|ask_qty`）
- 启动时从最新快照恢复订单簿（用 `bufio.Scanner` 逐行读取）
- 用 `os.Stat` 检查文件是否存在，用 `os.Remove` 清理过期快照（保留最近 10 个）

**扩展**：
- 用 JSON 格式替代文本格式（`json.Encoder` + `bufio.Writer` 流式编码）
- 用 `filepath.Walk` 遍历快照目录找到最新文件

---

#### 9-3：gRPC 风格请求体流处理

**难度**：⭐⭐⭐

**题目**：用 `io.MultiReader` 实现批量订单提交：

**要求**：
- 用户可以提交多个订单（用 `\n` 分隔的 JSON）
- 服务端用 `bufio.Reader` 流式读取每个 JSON 行
- 用 `io.ReadBytes('\n')` 或 `bufio.Scanner` 逐行解析
- 对每一行独立解析和验证，互不干扰
- 部分成功时返回成功和失败的订单列表
- 用 `io.Copy` 将响应流式写回客户端（不等待全部处理完再返回）

---

## 第十章 · Gin 框架核心与中间件

### 练习题

---

#### 10-1：撮合引擎 API 路由设计

**难度**：⭐

**题目**：设计撮合引擎的完整 API 路由：

```go
// 公开 API（无需认证）
GET  /api/v1/time                          → 服务器时间
GET  /api/v1/depth?symbol=BTCUSDT&limit=20 → 订单簿深度
GET  /api/v1/trades?symbol=BTCUSDT         → 成交历史
GET  /api/v1/ticker?symbol=BTCUSDT         → 24小时行情

// 私有 API（需要认证）
POST /api/v1/order          → 下单
DELETE /api/v1/order/:oid   → 撤单
GET  /api/v1/orders        → 当前挂单列表
GET  /api/v1/orders/history → 历史订单
GET  /api/v1/account/balance → 账户余额
GET  /api/v1/positions      → 持仓列表

// WebSocket
WS /ws?streams=btcusdt@depth@100ms,btcusdt@trade
```

**要求**：
- 用 `router.Group()` 实现分组（公开组 / 私有组 / WebSocket 组）
- JWT 认证中间件保护私有 API
- CORS 中间件（手写，不用库）
- 路由注册函数签名：`func RegisterRoutes(r *gin.Engine, engine *MatchingEngine)`

---

#### 10-2：自定义 Binding（限价单解析）

**难度**：⭐⭐⭐

**题目**：实现自定义的限价单 JSON Binding：

**要求**：
- 实现 `LimitOrderBinder`：`Bind(*http.Request, obj interface{}) error` 接口
- 支持解析以下格式：
  - 标准 JSON：`{"symbol":"BTCUSDT","price":65000,"qty":1.0,"side":"BUY"}`
  - 含注释的 JSON（允许末尾逗号等宽松格式）
- 用 `strings.Split` + `strconv.ParseFloat` 手动解析（不用 `json.Unmarshal`，作为练习）
- 在 Gin 中注册自定义 binding
- 实现 `ShouldBindLimitOrder` 辅助函数

---

#### 10-3：Gin 中间件全家桶

**难度**：⭐⭐

**题目**：实现以下手写中间件（不用第三方库）：

1. **Logger 中间件**：记录 method、path、status、latency、client_ip、trace_id、order_id（如果有）
2. **RateLimit 中间件**（滑动窗口）：基于 IP + UserID 的双重限流
   - IP 级别：100 req/s
   - UserID 级别：50 单/10s
   - 用 `sync.RWMutex` + `map[string]*windowCounter` 实现滑动窗口
3. **Recover 中间件**：捕获 panic，恢复后返回 JSON 500
4. **Timeout 中间件**：对每个请求设置 10s 超时（`context.WithTimeout`）
5. **Signature 中间件**：验证 API 签名（HMAC-SHA256），防止请求篡改

---

## 第十一章 · MySQL + Redis：持久化与缓存

### 练习题

---

#### 11-1：泛型 Repository 模式

**难度**：⭐⭐⭐

**题目**：使用 Go 1.18+ 泛型实现通用 Repository：

```go
type Repository[T any] struct {
    db    *sql.DB
    table string
}
```

**要求**：
- 实现 `Create(ctx context.Context, entity *T) error`
- 实现 `FindByID(ctx context.Context, id int64) (*T, error)`
- 实现 `FindAll(ctx context.Context, offset, limit int) ([]*T, int64, error)`
- 实现 `Update(ctx context.Context, entity *T) error`
- 实现 `Delete(ctx context.Context, id int64) error`
- 用 `reflect` 动态获取表名（从 struct 名推导：`Order` → `orders`）
- 用 `reflect` 动态从 `T` 获取字段名和字段值，构建 UPDATE SQL 语句
- 实例化 `UserRepo = Repository[User]`、`OrderRepo = Repository[Order]`

---

#### 11-2：Redis 缓存层

**难度**：⭐⭐⭐

**题目**：为撮合引擎添加 Redis 缓存：

**要求**：
- 缓存 K线数据（`kline:{symbol}:{interval}:{open_time}`，TTL 30秒）
- 缓存订单簿深度快照（`depth:{symbol}`，TTL 5秒）
- 缓存用户余额（`balance:{userID}:{asset}`，TTL 10秒）
- 实现缓存穿透保护：空结果也缓存（TTL 5秒）
- 实现缓存雪崩保护：TTL 加随机偏移（`ttl + rand(0, 300)`秒）
- 用 `Pipeline` 批量写入/读取多档深度数据
- 实现缓存击穿保护：用 `sync.RWMutex` 做单flight 模式（同一 key 同时只有 1 个请求去加载数据库）
- 用 `github.com/redis/go-redis/v9`（已有依赖），编写测试用 miniredis

---

## 第十二章 · GitHub 协作与 CI/CD

### 练习题

---

#### 12-1：Git 分支策略

**难度**：⭐

**题目**：设计并执行撮合引擎的 Git 分支策略：

```
main         → 生产分支（受保护）
develop      → 开发主分支
feature/*    → 功能分支
hotfix/*     → 热修复分支
```

**要求**：
- 从 `develop` 创建 `feature/matching-engine` 分支
- 实现限价单撮合逻辑（第五章练习的简化版）
- 用 `git rebase -i`（squash）合并到 develop
- 创建 Pull Request，设置 reviewers
- 创建 `hotfix/order-id-bug` 分支模拟紧急修复

---

#### 12-2：Git 钩子

**难度**：⭐

**题目**：编写项目 Git 钩子：

**要求**：
- `pre-commit`：运行 `gofmt -l` 检查格式，`go vet ./...` 检查错误
- `commit-msg`：强制 commit message 格式：`type(scope): description`
  - type：`feat`（新功能）、`fix`（bug修复）、`perf`（性能优化）、`test`（测试）、`refactor`（重构）、`docs`（文档）
  - 用正则验证：`^(feat|fix|perf|test|refactor|docs)\(.+\): .+`
- 将钩子文件放入 `.githooks/` 目录并用 `git config core.hooksPath .githooks` 配置

---

#### 12-3：GitHub Actions CI/CD

**难度**：⭐⭐

**题目**：编写完整的 CI/CD 工作流 `.github/workflows/ci.yml`：

**触发条件 1 — Push 到 PR / develop**：
- `go vet ./...` 静态检查
- `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
- `golangci-lint run`
- 上传覆盖率报告（`codecov/codecov-action`）

**触发条件 2 — 合并到 main**：
- 编译三个平台二进制（Linux / Windows / macOS）
- `docker build -t matchengine:{tag} .`
- 用 `aws-actions/amazon-ecr-login` + `docker push` 推送到 ECR
- 发送通知到 Slack（`SlackHQ/action-slack-notification`）

**触发条件 3 — 定时任务（每周一 03:00）**：
- `go mod tidy` + `go get -u ./...` 检查依赖更新
- 生成依赖变更报告并自动评论到 `develop` 分支的 PR

**触发条件 4 — PR 审查**：
- 自动评论：依赖变更列表、二进制大小对比

---

#### 12-4：语义化版本与 Changelog

**难度**：⭐⭐

**题目**：实现版本管理工具：

**要求**：
- 用 `git tag` 管理版本（`v1.0.0`）
- 用 `golang.org/x/mod/semver` 解析版本号
- 实现 `bump` 命令（patch/minor/major）
- 自动生成 CHANGELOG（按 `feat:` / `fix:` / `perf:` 分组 commit）
- 用 `strings.Builder` 高效拼接 CHANGELOG 文本

---

## 第十三章 · 综合实战：构建完整撮合引擎

### 13-1：项目结构设计

**难度**：⭐

**题目**：设计撮合引擎的完整项目结构并实际创建：

```
matchengine/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── matching/
│   │   ├── engine.go         # 撮合核心
│   │   ├── orderbook.go      # 订单簿
│   │   ├── order.go          # 订单结构
│   │   └── types.go          # 类型定义
│   ├── market/
│   │   ├── ticker.go         # 行情数据
│   │   └── kline.go          # K线聚合
│   ├── account/
│   │   ├── balance.go        # 余额管理
│   │   └── position.go       # 持仓管理
│   ├── ws/
│   │   ├── hub.go            # WebSocket Hub
│   │   ├── client.go         # 客户端管理
│   │   └── streams.go        # 流管理
│   ├── handler/
│   │   ├── order.go          # 订单接口
│   │   ├── market.go         # 市场数据接口
│   │   └── account.go        # 账户接口
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── ratelimit.go
│   │   ├── recovery.go
│   │   └── signature.go
│   ├── repository/
│   │   ├── order_repo.go
│   │   ├── trade_repo.go
│   │   └── balance_repo.go
│   ├── cache/
│   │   └── redis.go
│   └── router/
│       └── router.go
├── pkg/
│   ├── errors/
│   │   └── errors.go
│   ├── response/
│   │   └── response.go
│   ├── validator/
│   │   └── validator.go      # reflect 验证器
│   └── tools/
│       ├── bytesutil/        # bytes/bufio 工具
│       ├── strutil/          # strings 工具集
│       ├── unicodeutil/      # unicode 工具集
│       └── unsafeutil/       # unsafe 工具集
├── testdata/
│   └── orderbook_snapshot.sample
├── scripts/
│   └── init.sql
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

**要求**：
- 实际创建所有目录（空文件即可，结构即设计）
- 编写 `Makefile`：`make build`、`make test`、`make run`、`make lint`、`make docker`
- 用 `go mod init` 初始化（如尚未初始化）

---

#### 13-2：统一响应与错误处理

**难度**：⭐⭐

**题目**：实现统一的 API 响应格式：

```go
type Response struct {
    Code    int              `json:"code"`
    Msg     string           `json:"msg"`
    Data    interface{}      `json:"data,omitempty"`
    TraceID string           `json:"trace_id,omitempty"`
    Errors  []FieldError     `json:"errors,omitempty"`
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

**要求**：
- `Success(c *gin.Context, data interface{})`
- `Fail(c *gin.Context, code int, msg string)`
- `ValidationFail(c *gin.Context, errors []FieldError)`
- 在中间件中统一包装：后置处理（后置中间件模式）
- 用 `reflect` 从 `Response` struct 的 json tag 中提取字段信息

---

#### 13-3：撮合引擎核心实现

**难度**：⭐⭐⭐⭐

**题目**：实现限价单撮合引擎的核心逻辑：

**核心算法（价格-时间优先）：
1. 收到买单 → 从最低卖单（Best Ask）开始逐档扫描
2. 若 `bid_price >= ask_price` → 成交（数量取 min）
3. 若完全成交 → 继续扫描下一档
4. 若部分成交 → 更新双方订单的 Filled 字段
5. 若 `bid_price < ask_price` → 挂在买盘（Bid 档位）
6. 卖单同理（从最高买单开始扫描）

**要求**：
- `OrderBook.AddOrder(order *Order) ([]*Trade, error)`
- `OrderBook.CancelOrder(orderID int64) error`
- `OrderBook.GetDepth(limit int) (bids, asks []PriceLevel)`
- 用 `sync.RWMutex` 保护并发访问（读多写少场景）
- 用 `map[float64]*PriceLevel` 按价格索引（价格作为 key）
- 用 `sort.Slice` 或手动排序保持价格档位有序
- 实现完整的边界测试（参考第六章练习 6-1）
- 撮合触发时计算手续费（Maker 0.1%，Taker 0.2%），记录为 `Trade` 的一部分

**扩展**：
- 实现冰山订单（Iceberg Order）：大单仅显示部分数量，隐藏在档位后面
- 实现止盈止损单（Stop-Loss Order）：到达触发价后自动以市价单成交

---

#### 13-4：性能对比报告

**难度**：⭐⭐⭐

**题目**：对撮合引擎关键路径进行性能基准测试，生成 Markdown 报告：

**要求**：
- 对比 `strings.Builder` vs `fmt.Sprintf` vs `+` 拼接（成交记录日志）
- 对比 `bytes.Buffer` vs `[]byte` 手动拼接（二进制协议）
- 对比 `strconv.ParseFloat` vs 手写解析（订单参数）
- 对比 `StringToBytesUnsafe` vs `[]byte(s)`（日志热点路径）
- 对比 `sync.RWMutex` vs `sync.Mutex`（订单簿读写）
- 生成 Markdown 对比报告（表格格式）：

```markdown
| 场景 | 方法A | 方法B | 性能提升 | 内存节省 | 推荐 |
|------|-------|-------|---------|---------|------|
| 字符串拼接 | Builder | Sprintf | +35% | -40% | Builder |
```

---

## 学习顺序建议

```
第1-2周：第一章（strconv + strings）
第3周：第二章（bytes + bufio）
第4周：第三章（unicode + utf8）
第5-6周：第四章（reflect）+ 第五章（unsafe）
第7-8周：第六章（testing）+ 第七章（errors + fmt）
第9周：第八章（context）
第10周：第九章（net/http + io）
第11-12周：第十章（Gin 框架）+ 第十一章（MySQL + Redis）
第13周：第十二章（GitHub CI/CD）
第14周：第十三章（综合实战：完整撮合引擎）
```

## 验收标准

每个章节完成后，你的代码应满足：

1. **功能正确**：所有单元测试通过（`go test ./... -v`）
2. **代码规范**：通过 `gofmt` 和 `go vet`（`golangci-lint` 更佳）
3. **测试覆盖**：核心撮合逻辑覆盖率 ≥ 80%（`go test -coverprofile=cover.out`）
4. **有意义的 commit message**：符合 `type(scope): description` 格式
5. **Benchmark 有记录**：撮合引擎关键路径（撮合、订单解析、缓存）的性能数据有记录

---

## 附录：撮合引擎核心概念速查

### 订单类型

| 类型 | 说明 | 关键参数 |
|------|------|---------|
| 限价单（Limit） | 指定价格，挂在订单簿等待成交 | price, quantity |
| 市价单（Market） | 以最优对手价立即成交 | quantity |
| 冰山单（Iceberg） | 大单分批成交，仅显示部分 | display_qty, quantity |
| 止盈止损（Stop） | 触发后以市价单或限价单执行 | stop_price, side |

### 撮合优先级

```
价格优先 > 时间优先
买单：价格越高越优先（买一价 = 最高买价）
卖单：价格越低越优先（卖一价 = 最低卖价）
同价格：先提交的订单先成交（时间戳早的优先）
```

### 关键术语

- **Taker**：立即成交的订单（吃单）
- **Maker**：挂在订单簿等待成交的订单（挂单）
- **Bid/Ask**：买价/卖价
- **Depth**：订单簿深度（各档位的挂单量）
- **Spread**：买卖价差（Best Ask - Best Bid）
- **Notional**：订单名义价值（price × quantity）

---

*文档版本：v2.0（撮合引擎版）| 更新日期：2026-03-25*
