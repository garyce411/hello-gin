# Go 工具链全量修炼 · Gin 项目实战路线图

> **目标**：以一个真实项目 **"Go Learning Hub"** 为载体，系统掌握 Go 标准库的每一个工具，建立从"会用"到"用好"的完整能力闭环。
>
> **前提**：已安装 Go 1.21+，已安装 Git，了解基本的 HTTP 概念。
>
> **每个模块均为独立练习题，先独立思考，再查阅提示，最后参考答案。**

---

## 项目概览

### 我们要做什么

构建一个 **Go Learning Hub** —— 类似于一个极简的在线题库 / 代码片段管理平台，后端用 Gin 框架，支持：

- 用户注册登录（JWT）
- 题目的增删改查（CRUD）
- 用户提交代码并得到评测结果
- 代码片段收藏 / 笔记功能
- 管理后台（查看统计数据）
- API 文档自动生成

在这个过程中，**每个功能模块都刻意使用不同的标准库**，让你在实战中自然掌握工具链。

### 章节速查

| 章节 | 主题 | 核心库 |
|------|------|--------|
| 第一章 | 项目初始化与 strconv | `strconv` |
| 第二章 | 路由、路径与 strings | `strings` |
| 第三章 | 二进制数据与 bytes | `bytes`, `bufio` |
| 第四章 | Unicode 国际化与 runes | `unicode`, `utf8` |
| 第五章 | 动态反射与 reflect | `reflect` |
| 第六章 | 内存操作与 unsafe | `unsafe` |
| 第七章 | 单元测试与基准测试 | `testing` |
| 第八章 | 错误处理与 errors | `errors`, `fmt` |
| 第九章 | 并发上下文与 context | `context` |
| 第十章 | HTTP 客户端与 net/http | `net/http`, `io` |
| 第十一章 | Gin 框架核心与中间件 | `gin` |
| 第十二章 | Gin + MySQL + Redis | `database/sql`, `go-redis` |
| 第十三章 | GitHub 协作与 CI/CD | `git`, `gh` |
| 第十四章 | 性能优化综合实战 | 全部库综合 |

---

## 第一章 · strconv：字符串与数值的桥梁

### 背景

`strconv` 是 Go 中处理字符串与其他数据类型（bool、int、float、uint）之间转换的标准工具。在 Web 开发中，几乎所有从 HTTP 请求中获取的数字类参数（分页页码、ID、limit）都需要它。

### 核心 API

```go
// 字符串 → 其他类型
strconv.Atoi(s string) (int, error)
strconv.ParseInt(s string, base int, bitSize int) (int64, error)
strconv.ParseUint(s string, base int, bitSize int) (uint64, error)
strconv.ParseFloat(s string, bitSize int) (float64, error)
strconv.ParseBool(s string) (bool, error)

// 其他类型 → 字符串
strconv.Itoa(i int) string
strconv.FormatInt(i int64, base int) string
strconv.FormatUint(i uint64, base int) string
strconv.FormatFloat(f float64, fmt byte, prec, bitSize int) string
strconv.FormatBool(b bool) string
```

### 练习题

---

#### 1-1：Query 参数解析器

**难度**：⭐

**题目**：编写一个函数 `ParseQueryInt(c *gin.Context, key string, defaultVal int) (int, error)`，从 Gin 的 `c.Request.URL.Query()` 中读取 key 对应的值，如果不存在或解析失败则返回 defaultVal。

**要求**：
- 使用 `strconv.Atoi` 解析
- 错误只作为"降级"信号，不 panic
- 编写 3 个以上的单元测试（包含正常值、缺失 key、非法格式三种 case）

**扩展**：
- 扩展为 `ParseQueryFloat(key string, defaultVal float64) (float64, error)`
- 扩展为 `ParseQueryBool(key string, defaultVal bool) (bool, error)`

---

#### 1-2：分页器

**难度**：⭐⭐

**题目**：为题目列表接口设计分页参数解析。

**要求**：
- 参数：`page`（第几页，默认为 1）、`page_size`（每页条数，默认为 10，上限 100）
- 函数签名：`func ParsePagination(c *gin.Context) (page, pageSize, offset, limit int)`
- `pageSize` 超过 100 时自动裁剪为 100
- `page` 小于 1 时自动修正为 1
- 使用 `strconv.Atoi` + 手动校验，不能依赖第三方库

**扩展**：
- 返回总页数（需要查数据库 count）
- 支持 `cursor` 分页模式（基于上一页最后一条记录的 ID）

---

#### 1-3：ID 安全校验

**难度**：⭐⭐

**题目**：编写一个中间件 `ValidateIDParam(paramName string)`，从 URL 参数中读取 ID，进行严格校验：
- 不能为空
- 必须是正整数
- 不能超出 `int64` 范围
- 校验失败返回 400 Bad Request，并附带 JSON 错误信息

**要求**：
- 核心用 `strconv.ParseUint` + `strconv.FormatUint`
- 用 `errors.New` 构造自定义错误类型
- 中间件函数签名：`func ValidateIDParam(paramName string) gin.HandlerFunc`

**扩展**：
- 支持 UUID 格式校验（自学 `google/uuid` 或 `github.com/google/uuid`）
- 支持雪花 ID 格式（自学 `github.com/sony/sonyflake`）

---

## 第二章 · strings：文本处理的瑞士军刀

### 背景

`strings` 包提供了最常用的字符串操作：分割、拼接、大小写转换、裁剪、查找、替换等。在 Gin 项目中，路由匹配、请求头解析、模板渲染、日志处理都离不开它。

### 核心 API

```go
// 裁剪
strings.Trim(s, cutset string) string
strings.TrimSpace(s string) string
strings.TrimPrefix(s, prefix string) string
strings.TrimSuffix(s, suffix string) string

// 分割与拼接
strings.Split(s, sep string) []string
strings.SplitN(s, sep string, n int) []string
strings.Join(elems []string, sep string) string

// 大小写
strings.ToLower(s string) string
strings.ToUpper(s string) string
strings.ToTitle(s string) string
strings.Title(s string) string  // 每个单词首字母大写

// 查找
strings.Contains(s, substr string) bool
strings.HasPrefix(s, prefix string) bool
strings.HasSuffix(s, suffix string) bool
strings.Index(s, substr string) int
strings.LastIndex(s, substr string) int

// 替换
strings.Replace(s, old, new string, n int) string
strings.ReplaceAll(s, old, new string) string

// 统计
strings.Count(s, substr string) int
strings.Fields(s string) []string  // 按空白字符分割

// 拼接优化
strings.Builder
strings.Reader
```

### 练习题

---

#### 2-1：路由前缀与路径解析

**难度**：⭐

**题目**：编写一个函数 `NormalizePath(path string) string`，对 URL 路径做规范化处理：
- 去除首尾空白
- 去除末尾的 `/`（除非 path 就是 `/`）
- 多个连续斜杠替换为单个 `/`
- 转为小写

**要求**：
- 至少用 `strings.Trim`、`strings.ReplaceAll`、`strings.HasSuffix`、`strings.TrimSuffix` 四种方法
- 编写完整的表驱动测试（至少 8 个 case）

**示例**：
```
"/api//users///"   → "/api/users"
"/Admin/Users/"    → "/admin/users"
"  /api/v1/  "     → "/api/v1"
"/"                → "/"
```

---

#### 2-2：请求头安全过滤

**难度**：⭐⭐

**题目**：编写一个函数 `FilterHeaders(rawHeaders map[string]string) (safeHeaders map[string]string, dangerous bool)`，对 HTTP 请求头进行安全检查：
- 过滤掉 `X-Forwarded-For`、`X-Real-IP`、`Cookie` 等敏感头（要求保留但标记）
- 拒绝包含 `\r\n`（换行注入）的头
- 返回过滤后的安全头列表和是否有危险标记

**要求**：
- 用 `strings.Contains`、`strings.ToLower`、`strings.Split` 实现
- 用 `strings.Builder` 高效拼接日志输出
- 编写单元测试覆盖正常请求、注入攻击、多种敏感头

---

#### 2-3：搜索关键词高亮

**难度**：⭐⭐⭐

**题目**：编写一个函数 `Highlight(text, keyword string) string`，将文本中所有匹配的关键词用 `<mark>` 标签包裹。

**要求**：
- 搜索不区分大小写（用 `strings.ToLower` 归一化）
- 不替换掉已有的 `<mark>` 标签内的内容
- 关键词边界处理：不能切断 HTML 标签或 entity
- 示例：`Highlight("Go is great, golang too", "go")` → `"<mark>Go</mark> is great, <mark>Go</mark>lang too"`

**扩展**：
- 支持多个关键词（`[]string`）
- 支持正则表达式（引入 `regexp` 包自学，结合 `regexp.MatchString`）

---

#### 2-4：CSV 字段解析

**难度**：⭐⭐

**题目**：实现一个 CSV 行解析器 `ParseCSVLine(line string) ([]string, error)`，处理带引号和转义的 CSV 格式：
- 字段可被双引号包裹
- 引号内可以有逗号
- 双引号在引号内用 `""` 表示
- 支持自定义分隔符（默认逗号）

**要求**：
- 不使用 `encoding/csv`，手写实现
- 用 `strings.Split`、`strings.TrimSpace` 处理非引号部分
- 用状态机思维（引号内 / 引号外）处理引号部分
- 边界条件：空行、引号不闭合、连续的引号

---

## 第三章 · bytes 与 bufio：二进制与流的世界

### 背景

`bytes` 包操作 `[]byte`（字节切片），`bufio` 包提供带缓冲的读写。两者在处理 HTTP 请求体（二进制或流式数据）、文件操作、网络协议解析时不可或缺。

### 核心 API

```go
// bytes
bytes.NewBuffer(buf []byte) *Buffer
bytes.NewReader(buf []byte) *Reader
bytes.Buffer{}  // 可变长缓冲
bytes.Join(s [][]byte, sep []byte) []byte
bytes.Split(s, sep []byte) [][]byte
bytes.Contains(b, subslice []byte) bool
bytes.Equal(a, b []byte) bool
bytes.Compare(a, b []byte) int
bytes.Runes(b []byte) []rune  // bytes → runes

// bufio
bufio.NewReader(r io.Reader) *Reader
bufio.NewWriter(w io.Writer) *Writer
bufio.NewScanner(r io.Reader) *Scanner  // 按行/词扫描
```

### 练习题

---

#### 3-1：请求体内存复用（Buffer Pool）

**难度**：⭐⭐

**题目**：实现一个 `BodyPool`，复用 `bytes.Buffer` 减少 GC 压力：
- `Get() *bytes.Buffer`：获取一个 buffer，自动 Reset
- `Put(b *bytes.Buffer)`：归还 buffer（超过池子上限时丢弃）
- 在 Gin 中间件里用 Pool 包装请求体读取
- 编写 Benchmark 对比使用 Pool vs 不使用 Pool 的性能差异

**要求**：
- 使用 `sync.Pool`（第十章会深入）
- 用 `bytes.NewBuffer`、`b.Reset()`、`b.Write` 等操作
- Benchmark 用 `testing.B`，至少运行 10000 次迭代

**扩展**：
- 支持 Pool 的容量上限（超过 64KB 的 buffer 不回收）

---

#### 3-2：分块传输编码（Chunked Encoding）

**难度**：⭐⭐⭐

**题目**：用 `bufio.Reader` 实现一个简单的 HTTP 分块响应：
- 模拟一个大文件的流式读取
- 按固定大小 chunk（比如 4KB）读取文件
- 每个 chunk 前面输出十六进制长度
- 最后一个 chunk 输出 `0\r\n\r\n` 表示结束

**要求**：
- 用 `bufio.NewReaderSize` 和 `bufio.NewWriterSize`
- 用 `bytes.Buffer` 组装 chunk 头
- 参考 HTTP/1.1 Transfer-Encoding: chunked 规范
- 用 `net/http/httptest` 发起测试请求验证格式正确性

**示例**：
```
5\r\n
hello\r\n
0\r\n
\r\n
```

---

#### 3-3：二进制协议解析

**难度**：⭐⭐⭐⭐

**题目**：实现一个极简的 Redis RESP 协议解析器（只支持简单字符串和整数）：
- RESP 协议格式：`+OK\r\n`、`-ERR\r\n`、`:100\r\n`、`-1`（null）、Bulk String
- 用 `bytes.Reader` 或 `bufio.Reader` 从 `[]byte` 读取
- 返回解析后的值和下一个读取位置

**要求**：
- 函数签名：`func ParseRESP(data []byte) (interface{}, int, error)`
- 支持 `+`（简单字符串）、`-`（错误）、`:`（整数）、`$`（Bulk String）
- 用 `bytes.SplitN`、`bytes.HasPrefix` 判断类型
- 完整测试所有 RESP 类型

**扩展**：
- 支持 Array 类型（`*`开头）
- 参考：https://redis.io/docs/reference/protocol-spec/

---

## 第四章 · unicode 与 utf8：Unicode 全攻略

### 背景

Go 中 `string` 是 UTF-8 字节序列，`rune` 是 Unicode 码点（int32 的别名）。处理中文、日文、emoji、多语言场景时，必须正确处理 Unicode。`unicode` 包提供 Unicode 分类判断，`utf8` 包提供 UTF-8 编解码。

### 核心 API

```go
// rune 遍历
for i, r := range "hello世界" { }  // i 是字节索引，r 是 rune

// unicode 包
unicode.IsLetter(r rune) bool
unicode.IsDigit(r rune) bool
unicode.IsUpper(r rune) bool
unicode.IsLower(r rune) bool
unicode.IsSpace(r rune) bool
unicode.IsPunct(r rune) bool
unicode.IsMark(r rune) bool  // 变音符号
unicode.ToUpper(r rune) rune
unicode.ToLower(r rune) rune
unicode.ToTitle(r rune) rune

// utf8 包
utf8.RuneCountInString(s string) int
utf8.DecodeRuneInString(s string) (r rune, size int)
utf8.DecodeRune(p []byte) (r rune, size int)
utf8.EncodeRune(p []byte, r rune) int
utf8.FullRune(p []byte) bool
utf8.Valid(p []byte) bool
utf8.ValidString(s string) bool
```

### 练习题

---

#### 4-1：字符分类统计

**难度**：⭐⭐

**题目**：编写 `AnalyzeText(text string) TextStats`，统计一段文本中各类字符的数量：

```go
type TextStats struct {
    TotalBytes    int
    TotalRunes    int
    Letters       int
    Digits         int
    Spaces         int
    ChineseChars   int  // CJK统一汉字
    Punct          int
    Emoji          int  // emoji 范围 U+1F300 - U+1F9FF
    MaxRuneLen     int  // 最长的 rune 序列（字节数）
}
```

**要求**：
- 用 `for i, r := range text` 遍历
- 用 `unicode.IsLetter`、`unicode.IsDigit`、`unicode.IsSpace` 判断
- 用 `unicode.SimpleFold` 检测中文
- 用 `utf8.RuneCount` 验证
- 至少 10 个单元测试用例（包含纯英文、纯中文、emoji 混合、空字符串等）

---

#### 4-2：Unicode 归一化（模拟 NFC/NFD）

**难度**：⭐⭐⭐

**题目**：实现一个 Unicode 归一化函数 `NormalizeString(s string, mode string) string`：
- `mode = "upper"`：全部转为大写（使用 `unicode.ToUpper` 逐 rune 处理）
- `mode = "lower"`：全部转为小写
- `mode = "title"`：每个单词首字母大写（需要判断单词边界）

**要求**：
- 不能直接用 `strings.ToUpper` / `strings.ToLower`（需要逐 rune 处理）
- 用 `unicode.SpecialCase` 处理特殊大小写映射（如 `İ` → `i̇`）
- 用 `utf8.EncodeRune` 将处理后的 rune 写回字节数组

**扩展**：
- 实现完整的 NFD 分解（将组合字符分离为基础字符 + 组合符号），可参考 `unicode/norm` 标准库

---

#### 4-3：UTF-8 验证与修复

**难度**：⭐⭐⭐

**题目**：编写一个 `ValidateAndFixUTF8(data []byte) ([]byte, error)` 函数：
- 用 `utf8.Valid` 检查数据是否有效 UTF-8
- 如果无效，找出所有无效字节序列的位置
- 用 `unicode.ReplacementChar`（U+FFFD）替换无效字节
- 返回修复后的数据和所有错误位置列表

**要求**：
- 用 `utf8.FullRune`、`utf8.DecodeRune` 手写验证逻辑
- 不能用 `strings.ToValidUTF8`（那是 Go 1.15 新加的）
- 构造各种无效 UTF-8 测试用例（截断、多字节首字节错误、过长编码等）

**扩展**：
- 实现 `utf8string` 包风格的链式调用：
  `NewUTF8String(data).ToUpper().TrimSpace().Reverse().String()`

---

## 第五章 · reflect：反射与动态世界

### 背景

`reflect` 包让你在运行时检查变量的类型和值，是构建通用工具（ORM、DI 容器、序列化库、验证器）的基石。但反射有性能开销，生产代码应谨慎使用。

### 核心 API

```go
reflect.TypeOf(i interface{}) reflect.Type
reflect.ValueOf(i interface{}) reflect.Value
reflect.Kind  // 如 reflect.Struct, reflect.Slice, reflect.Int

// 结构性检查
v.Kind() == reflect.Struct
v.NumField() int
v.Field(i int) reflect.Value
v.FieldByName(name string) reflect.Value
v.Type().Field(i).Tag  // 获取 struct tag

// 指针和接口
v.Elem() reflect.Value  // 解引用
v.Interface() interface{}  // 转回 interface{}
v.CanSet() bool  // 是否可写

// 动态调用
v.Method(i int).Call(args []reflect.Value)
v.FieldByIndex(index []int) reflect.Value  // 嵌套字段

// 切片和 map
v.Len() int
v.Index(i int) reflect.Value
v.MapKeys()
v.MapIndex(key reflect.Value) reflect.Value
```

### 练习题

---

#### 5-1：通用结构体验证器

**难度**：⭐⭐⭐

**题目**：用 `reflect` 实现一个结构体标签验证器：

```go
// 标签用法示例
type User struct {
    Name  string `validate:"required,min=2,max=50"`
    Age   int    `validate:"required,min=0,max=150"`
    Email string `validate:"required,email"`
    URL   string `validate:"url"`
}
```

**要求**：
- 实现 `ValidateStruct(s interface{}) error`
- 支持的验证规则：
  - `required`：非空（string 非空，int > 0，指针非 nil）
  - `min=X`：最小值 / 最小长度
  - `max=X`：最大值 / 最大长度
  - `email`：邮箱格式（简单正则）
  - `len=X`：精确长度
  - `oneof=a,b,c`：枚举
- 遍历所有字段，用 `v.Field(i).Tag.Get("validate")` 读取标签
- 在 Gin 中间件中使用这个验证器

---

#### 5-2：通用 JSON 扁平化

**难度**：⭐⭐⭐

**题目**：用 `reflect` 实现一个函数 `FlattenJSON(v interface{}, prefix string) map[string]interface{}`，将嵌套的 struct/map 扁平化为单层 map：

**要求**：
- 嵌套 struct 字段用点号连接：`user.address.city`
- 嵌套 slice 用索引：`items.0.name`、`items.1.name`
- 嵌套 map 用 key：`config.db.host`
- 支持递归（struct 里嵌套 struct、slice 里嵌套 map）
- 用 `v.Kind()` 判断类型分支（Struct、Map、Slice、Primitive）

**示例**：
```go
type Address struct { City string `json:"city"` }
type User struct { Name string; Addr Address }
FlattenJSON(User{Name: "Alice", Addr: Address{City: "Beijing"}})
// → {"Name": "Alice", "Addr.City": "Beijing"}
```

---

#### 5-3：动态结构体创建

**难度**：⭐⭐⭐⭐

**题目**：实现一个 `DynamicBuilder`，能根据 map 创建任意类型的 struct 实例：

```go
func BuildStruct(data map[string]interface{}, target interface{}) error
```

**要求**：
- `target` 是指向 struct 的指针，用 `reflect.TypeOf` 获取类型
- 遍历 struct 的每个字段，从 `data` 中按字段名或 JSON tag 取值
- 自动做类型转换（如 `float64` → `int`，`string` → `int`）
- 处理 `time.Time`（RFC3339 格式字符串 → Time）
- 处理 `url.Values`（Gin 的 form 解析结果）到 struct 的映射

**扩展**：
- 实现 `ToMap`（struct → map），利用 `v.Field(i).Interface()`
- 在 Gin 的 `ShouldBind` 自定义实现中集成（`customBinding`）

---

## 第六章 · unsafe：内存操作的禁区

### 背景

`unsafe` 允许绕过 Go 的类型安全，直接操作内存。使用场景极少（性能关键代码、C 互操作、底层数据结构），但理解它是理解 Go 内存模型的关键。**警告：在生产代码中非必要不使用。**

### 核心 API

```go
// 指针转换
unsafe.Pointer(p *T) unsafe.Pointer
unsafe.Add(ptr unsafe.Pointer, n uintptr) unsafe.Pointer  // Go 1.17+
unsafe.Offsetof(st.field) uintptr
unsafe.Sizeof(v T) uintptr

// 特殊用法
// 将 []byte 转为 string（零拷贝）
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
// 将 string 转为 []byte（零拷贝）
func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
```

### 练习题

---

#### 6-1：零拷贝字符串转换

**难度**：⭐⭐⭐

**题目**：实现两个函数：
- `StringToBytesUnsafe(s string) []byte`：不分配内存，将 string 转为 []byte
- `BytesToStringUnsafe(b []byte) string`：不分配内存，将 []byte 转为 string

**要求**：
- 使用 `unsafe.Pointer` 和指针类型转换实现
- 写 Benchmark 对比 unsafe 版本 vs `[]byte(s)` / `string(b)` 的性能
- Benchmark 至少运行 1,000,000 次
- 分析内存分配差异（用 `testing.B.ReportMetric` 报告 alloc/op）

**思考**：
- 转换后的 `[]byte` 是否可以修改？（答案是**不能**，会导致未定义行为）
- 什么场景下这种零拷贝是有价值的？（高并发短字符串处理）

---

#### 6-2：结构体内存布局分析

**难度**：⭐⭐⭐⭐

**题目**：编写一个 `StructLayout(t reflect.Type)` 函数，打印任意 struct 的内存布局：

```go
type LayoutInfo struct {
    FieldName  string
    TypeName   string
    Offset     uintptr
    Size       uintptr
    Alignment  uintptr
    IsPadded   bool
}
```

**要求**：
- 用 `reflect.TypeOf` 获取类型，用 `t.NumField()` 和 `t.Field(i)` 遍历
- 用 `unsafe.Offsetof`、`unsafe.Sizeof` 获取字段偏移和大小
- 判断是否有内存空洞（padding）
- 打印类似 C 语言 `sizeof` 报告的格式

**扩展**：
- 分析标准库中 `time.Time`、`sync.Mutex` 等结构的内存布局
- 验证是否符合 Go 内存对齐规则

---

#### 6-3：slice 头结构探秘

**难度**：⭐⭐⭐⭐

**题目**：通过 `unsafe` 手动构建一个 slice，深入理解 slice 的内部结构：

```go
type SliceHeader struct {
    Data uintptr  // 指向底层数组的指针
    Len  int      // 长度
    Cap  int      // 容量
}
```

**要求**：
- 用 `unsafe.Pointer` 将 `[]int{1,2,3}` 的地址强转为 `*SliceHeader`
- 验证 Data、Len、Cap 的值
- 通过修改 SliceHeader 手动 append 一个元素（不使用 append 关键字）
- 理解 slice 和 array 的关系

**警告**：
- 这个练习仅用于学习，不要在生产代码中这样做

---

## 第七章 · testing：质量保障体系

### 背景

Go 的 `testing` 包是标准测试框架，支持单元测试、性能测试（benchmark）、示例测试、子测试、子基准测试、Fuzzing 测试。Go 1.21 引入了 `testing.Testing()` 和 `slices`/`maps` 标准库。

### 核心 API

```go
// 基础测试
func TestXxx(t *testing.T)
func BenchmarkXxx(b *testing.B)
func ExampleXxx()  // 示例测试，输出与注释对比

// 子测试
t.Run(name string, f func(t *testing.T))
t.Parallel()  // 并行运行子测试

// 断言风格
t.Fatal / t.Fatalf  // 测试失败即停止
t.Error / t.Errorf  // 测试失败继续执行
t.Skip / t.Skipf    // 跳过测试

// Benchmark
b.ReportMetric(float64, "ns/op")    // 自定义指标
b.ResetTimer()
b.RunParallel(func(pb *testing.PB){})

// Fuzzing (Go 1.18+)
func FuzzXxx(f *testing.F)
```

### 练习题

---

#### 7-1：表驱动测试重构

**难度**：⭐

**题目**：将以下散乱的测试代码重构为表驱动测试：

```go
// 原代码（伪代码，不要真的写这种风格）
func TestAdd(t *testing.T) {
    if add(1, 2) != 3 { t.Fatal("fail") }
    if add(0, 0) != 0 { t.Fatal("fail") }
    if add(-1, 1) != 0 { t.Fatal("fail") }
    // ... 更多 case 线性叠加
}
```

**要求**：
- 重构为标准表驱动测试格式（`cases []struct{...}` + `for _, tc := range cases`）
- 每个 case 包含：`name`（子测试名）、`input`、期望值`、`wantErr`（是否期望错误）
- 用 `t.Run(tc.name, func(t *testing.T) { ... })` 创建子测试
- 添加至少 10 个 case（边界值、空输入、负数、超大数等）

---

#### 7-2：Mock 接口测试

**难度**：⭐⭐

**题目**：为以下接口编写 Mock，实现完整的单元测试：

```go
type UserRepository interface {
    FindByID(id int64) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int64) error
    List(page, pageSize int) ([]*User, int64, error)
}
```

**要求**：
- 手写 Mock 结构体（不需要第三方 Mock 库）
- 在 Mock 中用 `sync.RWMutex` 保护 map 模拟内存存储
- 实现"期望失败"的 Mock（通过构造函数参数注入）
- 用子测试验证每个方法
- 用 `t.Run("FindByID/not_found", ...)` 的子测试组织

**扩展**：
- 用 `github.com/stretchr/testify/assert` 的 `assert.Equal` 等断言

---

#### 7-3：Benchmark 性能分析

**难度**：⭐⭐⭐

**题目**：对第一章的分页器函数进行完整的性能分析：

**要求**：
- 编写 `BenchmarkParsePagination`（至少 10,000,000 次迭代）
- 用 `b.ReportAllocs()` 报告内存分配
- 对比三种实现：
  1. 纯 `strconv.Atoi`（baseline）
  2. 手动手写解析（`for` 循环 + 乘法）
  3. 正则表达式方案
- 使用 `go test -bench=. -benchmem -cpuprofile=cpu.out` 生成 CPU profile
- 用 `go tool pprof` 分析热点（命令行或可视化）

**扩展**：
- 用 `go test -fuzz` 对分页参数做模糊测试
- 用 `stretchr/testify/require` 的 `Eventually` 做异步测试

---

#### 7-4：Golden File 测试

**难度**：⭐⭐

**题目**：实现一个 JSON 响应的 Golden File 测试框架：

**要求**：
- 第一次运行测试时，如果 golden 文件不存在，则自动生成（写文件）
- 后续运行测试时，对比实际输出与 golden 文件内容
- golden 文件命名：`testdata/<testname>.golden`
- 用 `os.WriteFile` / `os.ReadFile` 或 `os.Create`
- 用 `filepath.Join` 构建路径
- 处理 `encoding/json` 输出的格式化（`json.MarshalIndent`）

---

## 第八章 · errors 与 fmt：错误处理艺术

### 背景

Go 1.13+ 引入的 `errors` 包（`errors.Is`、`errors.As`）配合 `fmt.Errorf` 的 `%w` 包装，实现了错误链。理解错误类型而非错误消息是做防御性编程的关键。

### 核心 API

```go
errors.New(s string) error
errors.Is(err, target error) bool
errors.As(err error, target interface{}) bool
errors.Join(errs ...error) error  // Go 1.20+

fmt.Errorf("...: %w", err)  // 包装错误
%v %+v %q %s %d  // 各格式化动词
fmt.Sprintf / fmt.Printf / fmt.Println
```

### 练习题

---

#### 8-1：自定义错误类型体系

**难度**：⭐⭐

**题目**：为题库系统设计一套分层错误类型：

```go
// 顶层接口
type Coder interface {
    Code() int  // HTTP 状态码映射
}

// 各层错误
type ValidationError struct { Field, Msg string }
type NotFoundError struct { Resource string; ID interface{} }
type UnauthorizedError struct{}
type InternalError struct{ Cause error }
type ConflictError struct{ Msg string }
```

**要求**：
- 每个错误类型实现 `Error() string` 和 `Coder` 接口
- 用 `errors.New` 或自定义构造函数创建
- 用 `errors.As` 在业务层按类型处理
- 在 Gin 中间件中统一处理错误（根据 `Coder.Code()` 返回对应 HTTP 状态码）

---

#### 8-2：错误日志与堆栈

**难度**：⭐⭐⭐

**题目**：实现一个错误日志中间件，记录每个请求的错误信息：

**要求**：
- 捕获所有 panic，恢复后记录错误和堆栈
- 用 `runtime.Callers` + `runtime.Stack` 获取堆栈信息
- 用自定义 JSON 格式输出：`{"time","level","error","stack","request_id"}`
- 支持不同日志级别（Debug / Info / Warn / Error / Fatal）
- 用 `fmt.Sprintf` + `strings.Builder` 高效拼接

**扩展**：
- 使用 `log/slog`（Go 1.21+ 结构化日志）重构

---

## 第九章 · context：并发取消与超时

### 背景

`context` 是 Go 并发编程的核心，用于：
- 请求作用域的取消信号（用户断开连接 → 取消数据库查询）
- 超时控制（SQL 查询超过 5 秒自动终止）
- 传递请求级别的值（trace_id、user_id）

### 核心 API

```go
context.Background() context.Context
context.TODO() context.Context
context.WithCancel(parent) (ctx, cancel func)
context.WithTimeout(parent, d time.Duration) (ctx, cancel func)
context.WithDeadline(parent, t time.Time) (ctx, cancel func)
context.WithValue(parent, key, val) context.Context
ctx.Err()  // 错误原因：Canceled 或 DeadlineExceeded
```

### 练习题

---

#### 9-1：超时控制的数据库查询

**难度**：⭐⭐

**题目**：模拟一个"用户画像查询"场景，需要并发查询三个数据源（用 goroutine 模拟）：
- 用户基本信息（耗时 100ms）
- 用户行为日志（耗时 200ms）
- 用户偏好推荐（耗时 150ms）

**要求**：
- 总超时时间 500ms
- 任意一个超时则整体失败
- 用 `context.WithTimeout` + `select` + `ctx.Done()` 实现
- 用 `sync.WaitGroup` 等待所有 goroutine 完成
- 返回最先返回的数据（"竞速"）

**扩展**：
- 实现"所有数据都成功才返回"（类似 `Promise.all`）
- 实现"最多等待 N 个"（用 channel 限流）

---

#### 9-2：请求链路追踪

**难度**：⭐⭐⭐

**题目**：实现一个带 trace_id 的请求链路：

**要求**：
- Gin 中间件：从请求头 `X-Request-ID` 读取 trace_id，如果没有则生成 UUID
- 将 trace_id 存入 `context.WithValue`
- 在日志、数据库操作、HTTP 客户端调用中通过 `FromContext` 获取 trace_id
- 打印结构化日志：`{"trace_id":"xxx","event":"db_query","duration_ms":12}`

**扩展**：
- 实现 OpenTelemetry 风格的 span 嵌套（自学 `go.opentelemetry.io/otel`）

---

## 第十章 · net/http、io、os：I/O 全景

### 背景

标准库 `net/http` 不仅是 Gin 的底层依赖，也是做 HTTP 客户端、文件服务、微服务通信的工具箱。`io`/`os` 包提供底层 I/O 抽象。

### 核心 API

```go
// net/http 服务端
http.HandleFunc / http.Handle
http.ListenAndServe(addr string, handler Handler)
http.Server{Addr, Handler, ReadTimeout, WriteTimeout}
http.ResponseWriter, http.Request
http.StatusOK, http.StatusBadRequest ...

// net/http 客户端
http.Get(url) (*http.Response, error)
http.NewRequest(method, url, body) (*http.Request, error)
http.Client{Timeout, Transport}
http.DefaultClient

// io
io.ReadAll(r io.Reader) ([]byte, error)
io.Copy(w io.Writer, r io.Reader) (int64, error)
io.ReadFull(r io.Reader, buf []byte) (int, error)
io.LimitReader(r io.Reader, n int64) io.Reader
io.MultiReader(m ...io.Reader) io.Reader
io TeeReader(r io.Reader, w io.Writer) io.Reader

// os
os.Open(name string) (*os.File, error)
os.ReadFile(name string) ([]byte, error)
os.WriteFile(name string, data []byte, perm FileMode) error
os.MkdirAll(path string, perm FileMode) error
```

### 练习题

---

#### 10-1：HTTP 客户端封装

**难度**：⭐⭐

**题目**：实现一个类型安全的 HTTP 客户端：

```go
type HTTPClient struct {
    BaseURL    string
    Client     *http.Client
    Headers    map[string]string
}
```

**要求**：
- 实现 `Get`、`Post`、`Put`、`Delete` 方法
- 自动添加 `Content-Type: application/json`
- 自动注入 `trace_id`（从 context）
- 实现自动重试（GET 请求失败时最多重试 3 次，间隔 100ms/200ms/400ms）
- 用 `httptest.NewServer` 启动测试服务器，写集成测试

---

#### 10-2：文件分片上传与断点续传

**难度**：⭐⭐⭐⭐

**题目**：实现一个分片文件上传服务：

**要求**：
- 客户端：将文件按 1MB 分片，逐片上传，携带 shard_index 和 total_shards
- 服务端：接收分片，写入临时目录
- 所有分片接收完毕后，合并为完整文件
- 支持断点续传：客户端可以查询已接收的分片列表，只上传缺失的分片
- 用 `os.Create`/`os.Write`/`os.Open` 实现分片读写

**扩展**：
- 用 `filepath.Join` 和 `os.MkdirAll` 管理上传目录
- 实现 SHA256 分片校验

---

#### 10-3：流式响应（Server-Sent Events）

**难度**：⭐⭐⭐

**题目**：用 `net/http`（不用 Gin）实现一个 SSE（Server-Sent Events）接口：

**要求**：
- HTTP header：`Content-Type: text/event-stream`
- 每秒向客户端推送一条消息（当前时间戳）
- 消息格式：`data: {"time":"2024-01-01T12:00:00Z"}\n\n`
- 客户端断开连接时正确关闭 goroutine（用 `context` + `select`）
- 用 `httptest.NewServer` 测试连接生命周期

---

## 第十一章 · Gin 框架核心

### 练习题

---

#### 11-1：路由分组与中间件链

**难度**：⭐

**题目**：设计题库系统的 API 路由结构：

```go
// 路由分组
/                          → 公开
  /auth/register           → POST
  /auth/login              → POST

/api/v1/                   → 需要认证
  /users                   → GET (列表), POST (创建)
  /users/:id               → GET, PUT, DELETE
  /problems                → GET (列表), POST (创建)
  /problems/:id            → GET, PUT, DELETE
  /problems/:id/submit     → POST (提交代码)
  /problems/:id/testcases  → GET (获取测试用例)
  /submissions/:id         → GET (查看评测结果)
```

**要求**：
- 用 `router.Group()` 实现分组
- 用 JWT 中间件保护私有路由
- 用 `CORS` 中间件（手写，不用库）
- 路由注册函数签名：`func RegisterRoutes(r *gin.Engine, srv *Server)`

---

#### 11-2：自定义 Binding

**难度**：⭐⭐⭐

**题目**：实现自定义的 `FormBinding`（支持 `application/x-www-form-urlencoded`）和 `TOML Binding`：

**要求**：
- 实现 `Bind(*http.Request, obj interface{}) error` 接口
- TOML Binding：解析 `application/x-toml` 请求体
- 在 Gin 的 `ShouldBind` 中注册自定义 binding
- 用 `reflect` + `strings.Split` + `strconv` 手动实现（不用第三方库）

---

#### 11-3：中间件系统

**难度**：⭐⭐

**题目**：编写以下 Gin 中间件（全部手写，不用第三方库）：

1. **Logger 中间件**：记录每个请求的 method、path、status_code、latency、client_ip、trace_id
2. **RateLimiter 中间件**：基于 IP 的限流（滑动窗口算法，使用 `sync.RWMutex` + `map`）
3. **Recover 中间件**：捕获 panic，返回 JSON 500 错误
4. **Timeout 中间件**：对每个请求单独设置超时（`context.WithTimeout`）

**要求**：
- 每个中间件独立，接口签名：`func(...) gin.HandlerFunc`
- 中间件可组合（`r.Use(m1, m2, m3)`）
- 编写测试验证每个中间件的行为

---

## 第十二章 · 数据库与缓存

### 练习题

---

#### 12-1：泛型 Repository 模式

**难度**：⭐⭐⭐

**题目**：使用 Go 1.18+ 泛型实现一个通用的 Repository：

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
- 使用 `reflect` 动态获取表名（从 struct 名推导）
- 在 Gin 中注册 UserRepository 和 ProblemRepository

---

#### 12-2：Redis 缓存层

**难度**：⭐⭐⭐

**题目**：为题库接口添加 Redis 缓存：

**要求**：
- 用 `github.com/redis/go-redis/v9`
- 缓存题目详情（`problem:{id}`，TTL 5 分钟）
- 缓存题目列表（`problems:list:page:{page}:size:{size}`，TTL 1 分钟）
- 实现缓存穿透保护（空结果也缓存，但 TTL 短）
- 实现缓存雪崩保护（TTL 加随机偏移）
- 用 `Pipeline` 批量写入
- 编写完整的单元测试（可以用 miniredis）

---

## 第十三章 · GitHub 协作与 CI/CD

### 背景

这一章不是 Go 代码练习，而是 GitHub 协作流程和 CI/CD 的实战训练。在实际项目中，代码质量和协作规范同样重要。

### 练习题

---

#### 13-1：Git 分支策略与工作流

**难度**：⭐

**题目**：为题库系统设计 Git 分支策略并执行：

**要求**：
- 创建以下分支：`main`、`develop`、`feature/xxx`、`hotfix/xxx`
- 从 `develop` 创建 `feature/strconv-validator` 分支
- 实现练习 1-3 的功能
- 用 `git rebase -i`（squash）合并到 develop
- 创建 Pull Request，设置 required reviewers（模拟代码审查）

---

#### 13-2：.gitignore 与 Git 钩子

**难度**：⭐

**题目**：编写项目的 `.gitignore` 和自定义 Git 钩子：

**要求**：
- `.gitignore` 忽略：二进制文件、`.env`（但保留 `.env.example`）、IDE 配置、`vendor/`（如果不用 Go Modules）、测试覆盖率报告
- 实现 `pre-commit` 钩子：运行 `go fmt` 检查（`gofmt -l`）
- 实现 `commit-msg` 钩子：强制 commit message 格式：`type(scope): description`
  - type: `feat`、`fix`、`docs`、`test`、`refactor`、`chore`
  - 用正则表达式验证
- 将钩子文件放入 `.git/hooks/`（或用 `git config core.hooksPath`）

---

#### 13-3：GitHub Actions CI/CD

**难度**：⭐⭐

**题目**：为项目编写完整的 GitHub Actions 工作流：

**要求**：
创建 `.github/workflows/ci.yml`：

1. **Push 到 PR 时**（自动触发）：
   - `go vet` 静态检查
   - `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
   - 上传覆盖率报告到 Codecov（`codecov/codecov-action`）

2. **合并到 main 时**：
   - 编译二进制：`GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/server`
   - 编译 Windows 版本：`GOOS=windows`
   - 编译 macOS 版本：`GOOS=darwin`
   - 用 `actions/upload-artifact` 上传二进制文件

3. **定时任务**（每周一凌晨 3 点）：
   - `go mod tidy` 检查依赖更新
   - 生成依赖报告（`go list -m all`）

4. **在 PR 中自动评论**：
   - 使用 `go list -f '{{.ImportPath}} {{.Version}}' -m all` 列出依赖版本
   - 自动评论到 PR 中，告知依赖变更情况

**扩展**：
- 添加 Docker 构建步骤（`docker build` + `docker push`）
- 添加 `golangci-lint` 高级检查

---

#### 13-4：语义化版本与 Changelog 自动生成

**难度**：⭐⭐

**题目**：实现基于 Git Tag 的语义化版本管理：

**要求**：
- 使用 `git tag` 管理版本（`v1.0.0`）
- 使用 `github.com/Masterminds/semver` 或 `golang.org/x/mod/semver` 解析版本号
- 实现 `bump` 命令：
  - `bump patch`：1.0.0 → 1.0.1
  - `bump minor`：1.0.0 → 1.1.0
  - `bump major`：1.0.0 → 2.0.0
- 用 `git tag` + `git push` 自动打标签
- 自动生成 CHANGELOG（从 commit message 中提取 feat/fix/docs，按类型分组）

---

## 第十四章 · 综合实战：构建完整的题库系统

### 背景

前面的每一章都是一个独立的练习点。本章要求将所有知识点串联起来，构建一个真正可以运行的题库系统。

### 练习题

---

#### 14-1：项目结构设计（架构图）

**难度**：⭐

**题目**：设计 Gin 题库系统的项目结构，并实际创建：

```
go-learning-hub/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── user.go
│   │   ├── problem.go
│   │   ├── submission.go
│   │   └── auth.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   ├── ratelimit.go
│   │   └── recovery.go
│   ├── model/
│   │   ├── user.go
│   │   ├── problem.go
│   │   └── submission.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   └── problem_repo.go
│   ├── service/
│   │   ├── user_svc.go
│   │   ├── problem_svc.go
│   │   └── judge_svc.go  # 代码评测服务
│   ├── validator/
│   │   └── validator.go  # reflect 验证器
│   └── router/
│       └── router.go
├── pkg/
│   ├── errors/
│   │   └── errors.go
│   ├── response/
│   │   └── response.go   # 统一 JSON 响应
│   └── tools/
│       └── bytesconv/    # unsafe 零拷贝
│       └── strutil/      # strings 工具集
│       └── unicodeutil/  # unicode 工具集
├── testdata/
├── scripts/
│   ├── migrate.sql
│   └── seed.sql
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
└── Makefile
```

**要求**：
- 实际创建所有目录和文件（空文件即可）
- 编写 `Makefile`，包含：`make build`、`make test`、`make run`、`make clean`、`make lint`
- 用 `go mod init` 初始化项目

---

#### 14-2：统一响应与错误处理

**难度**：⭐⭐

**题目**：设计并实现统一的 API 响应格式：

**要求**：
```go
// 统一响应格式
type Response struct {
    Code    int         `json:"code"`     // 业务码（0=成功）
    Message string      `json:"message"`  // 消息
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"trace_id,omitempty"`
    Errors  []FieldError `json:"errors,omitempty"`  // 字段级错误
}
```

- 实现 `Success(c *gin.Context, data interface{})`
- 实现 `Fail(c *gin.Context, code int, message string)`
- 实现 `ValidationFail(c *gin.Context, errors []FieldError)`
- 用 `reflect` 从 `Response` struct 的 json tag 中提取字段名
- 在中间件中统一包装所有响应（后置处理）

---

#### 14-3：代码评测沙箱

**难度**：⭐⭐⭐⭐

**题目**：实现一个极简的代码评测服务（不涉及真实安全沙箱，仅练习数据结构）：

**要求**：
- 题目模型包含：输入示例、输出示例、隐藏测试用例（JSON 存储在 MySQL 的 `text` 字段中，用 `bytes.Buffer` + `json.Marshal` 序列化）
- 提交代码后，读取测试用例（用 `os.ReadFile`），用正则表达式模拟运行结果（练习 2-3 的 `Highlight`）
- 用 `context.WithTimeout` 设置评测超时（5 秒）
- 用 `reflect` 动态调用不同语言的"评测函数"（实际上只模拟返回）
- 评测结果：Pending → Running → Accepted / Wrong Answer / Time Limit Exceeded

**扩展**：
- 引入 `os/exec` 执行真实的 Python/Java 代码（非常危险，仅供学习）

---

#### 14-4：性能对比报告

**难度**：⭐⭐⭐

**题目**：对项目中的关键路径进行性能基准测试，生成对比报告：

**要求**：
- 对比 `strings.Builder` vs `fmt.Sprintf` vs `+` 拼接（练习 2-1 涉及）
- 对比 `bytes.Buffer` vs `[]byte` 手动拼接
- 对比 `strconv.Atoi` vs 纯手写解析
- 对比 `StringToBytesUnsafe` vs `[]byte(s)`（练习 6-1 涉及）
- 生成 Markdown 格式的对比报告表格

**格式**：
```markdown
| 方法 | 每次操作耗时 | 内存分配 | 适用场景 |
|------|------------|---------|---------|
| ... | ... | ... | ... |
```

---

## 学习顺序建议

```
第1-2周：第一章 + 第二章（strconv + strings）
第3周：第三章（bytes） + 第四章（unicode）
第4-5周：第五章（reflect） + 第六章（unsafe）
第6-7周：第七章（testing）
第8-9周：第八章（errors） + 第九章（context）
第10周：第十章（net/http）
第11-12周：第十一章（Gin 框架）
第13周：第十二章（数据库） + 第十三章（GitHub）
第14周：第十四章（综合实战）
```

## 验收标准

每个章节完成后，你的代码应满足：

1. **功能正确**：所有单元测试通过（`go test ./... -v`）
2. **代码规范**：通过 `gofmt` 和 `go vet`（`golangci-lint` 更佳）
3. **测试覆盖**：核心函数覆盖率 ≥ 70%（`go test -coverprofile`）
4. **有意义的 commit message**：符合 `type(scope): description` 格式
5. **Benchmark 有记录**：关键路径的性能数据有记录（在 README 或独立文件中）

---

*文档版本：v1.0 | 更新日期：2026-03-25*
