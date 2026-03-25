package chapter01

import (
	"strings"
)

// QuoteAssetWhitelist 定义允许的计价资产白名单
var QuoteAssetWhitelist = map[string]bool{
	"USDT": true,
	"BUSD": true,
	"BTC":  true,
	"ETH":  true,
}

// SymbolConfig 交易对的配置信息
type SymbolConfig struct {
	BaseAsset   string  // 基础资产，如 "BTC"
	QuoteAsset  string  // 计价资产，如 "USDT"
	PricePrec   int     // 价格精度（小数位数）
	QtyPrec     int     // 数量精度
	MinQty      float64 // 最小下单量
	MaxQty      float64 // 最大下单量
	MinNotional float64 // 最小成交额（price * qty）
}

// SymbolValidator 交易对验证器，管理所有支持的交易对
type SymbolValidator struct {
	symbols map[string]*SymbolConfig
}

// NewSymbolValidator 创建新的交易对验证器，并注册默认交易对
func NewSymbolValidator() *SymbolValidator {
	sv := &SymbolValidator{
		symbols: make(map[string]*SymbolConfig),
	}

	// 注册默认交易对
	defaultPairs := []*SymbolConfig{
		{BaseAsset: "BTC", QuoteAsset: "USDT", PricePrec: 2, QtyPrec: 4, MinQty: 0.0001, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "ETH", QuoteAsset: "USDT", PricePrec: 2, QtyPrec: 4, MinQty: 0.0001, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "BNB", QuoteAsset: "USDT", PricePrec: 2, QtyPrec: 2, MinQty: 0.01, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "SOL", QuoteAsset: "USDT", PricePrec: 2, QtyPrec: 2, MinQty: 0.01, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "ADA", QuoteAsset: "USDT", PricePrec: 4, QtyPrec: 1, MinQty: 1, MaxQty: 100000, MinNotional: 10},
		{BaseAsset: "DOGE", QuoteAsset: "USDT", PricePrec: 5, QtyPrec: 0, MinQty: 10, MaxQty: 10000000, MinNotional: 10},
		{BaseAsset: "BTC", QuoteAsset: "BUSD", PricePrec: 2, QtyPrec: 4, MinQty: 0.0001, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "ETH", QuoteAsset: "BUSD", PricePrec: 2, QtyPrec: 4, MinQty: 0.0001, MaxQty: 9000, MinNotional: 10},
		{BaseAsset: "ETH", QuoteAsset: "BTC", PricePrec: 4, QtyPrec: 4, MinQty: 0.001, MaxQty: 9000, MinNotional: 0.001},
	}

	for _, cfg := range defaultPairs {
		symbol := BuildSymbol(cfg.BaseAsset, cfg.QuoteAsset)
		sv.symbols[symbol] = cfg
	}

	return sv
}

// ParseSymbol 解析交易对符号为基础资产和计价资产
// 支持三种格式：
//   - 无分隔符："BTCUSDT" → ("BTC", "USDT")
//   - 斜杠分隔："BTC/USDT" → ("BTC", "USDT")
//   - 横杠分隔："BTC-USDT" → ("BTC", "USDT")
func ParseSymbol(symbol string) (base, quote string, err error) {
	// 去除首尾空白
	symbol = strings.TrimSpace(symbol)

	// 规范化为大写
	symbol = strings.ToUpper(symbol)

	// 检查空字符串
	if symbol == "" {
		return "", "", &SymbolError{Message: "symbol不能为空"}
	}

	// 查找分隔符：先检查斜杠，再检查横杠
	separator := ""
	sepIdx := -1

	slashIdx := strings.Index(symbol, "/")
	dashIdx := strings.Index(symbol, "-")

	// 如果同时存在两种分隔符，使用第一个
	if slashIdx != -1 && dashIdx != -1 {
		if slashIdx < dashIdx {
			separator = "/"
			sepIdx = slashIdx
		} else {
			separator = "-"
			sepIdx = dashIdx
		}
	} else if slashIdx != -1 {
		separator = "/"
		sepIdx = slashIdx
	} else if dashIdx != -1 {
		separator = "-"
		sepIdx = dashIdx
	}

	var parts []string
	if sepIdx == -1 {
		// 无分隔符：根据已知的计价资产后缀来猜测
		quoteCandidates := []string{"USDT", "BUSD", "BTC", "ETH"}
		for _, quoteCandidate := range quoteCandidates {
			if strings.HasSuffix(symbol, quoteCandidate) {
				sepIdx = len(symbol) - len(quoteCandidate)
				base = symbol[:sepIdx]
				quote = symbol[sepIdx:]
				parts = []string{base, quote}
				break
			}
		}
		if parts == nil {
			return "", "", &SymbolError{
				Message: "无法解析symbol：未找到分隔符且无已知计价资产后缀",
				Symbol:  symbol,
			}
		}
	} else {
		// 有分隔符，使用SplitN切分为恰好2部分
		parts = strings.SplitN(symbol, separator, 2)
		if len(parts) != 2 {
			return "", "", &SymbolError{
				Message: "无效的symbol格式",
				Symbol:  symbol,
			}
		}
		base = parts[0]
		quote = parts[1]
	}

	// 去除各部分首尾空白
	base = strings.TrimSpace(base)
	quote = strings.TrimSpace(quote)

	// 校验基础资产不能为空
	if base == "" {
		return "", "", &SymbolError{Message: "基础资产不能为空", Symbol: symbol}
	}

	// 校验计价资产不能为空
	if quote == "" {
		return "", "", &SymbolError{Message: "计价资产不能为空", Symbol: symbol}
	}

	// 校验计价资产必须在白名单内
	if !QuoteAssetWhitelist[quote] {
		return "", "", &SymbolError{
			Message: "计价资产不在白名单内，允许值：USDT, BUSD, BTC, ETH",
			Symbol:  symbol,
		}
	}

	// 校验基础资产不能是已知的计价资产（防止BTCUSDT被解析为BTC/USDT）
	if QuoteAssetWhitelist[base] {
		return "", "", &SymbolError{
			Message: "基础资产不能是计价资产",
			Symbol:  symbol,
		}
	}

	// 校验基础资产只能包含字母
	for _, r := range base {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return "", "", &SymbolError{
				Message: "基础资产只能包含字母和数字",
				Symbol:  symbol,
			}
		}
	}

	return base, quote, nil
}

// BuildSymbol 将基础资产和计价资产拼接为标准格式
// 示例：BuildSymbol("BTC", "USDT") → "BTCUSDT"
func BuildSymbol(base, quote string) string {
	// 使用strings.Builder高效拼接
	var builder strings.Builder
	builder.Grow(len(base) + len(quote))
	builder.WriteString(strings.ToUpper(strings.TrimSpace(base)))
	builder.WriteString(strings.ToUpper(strings.TrimSpace(quote)))
	return builder.String()
}

// GetSymbolConfig 获取交易对的配置信息
func (sv *SymbolValidator) GetSymbolConfig(symbol string) (*SymbolConfig, error) {
	base, quote, err := ParseSymbol(symbol)
	if err != nil {
		return nil, err
	}

	// 在symbols映射中查找
	key := BuildSymbol(base, quote)
	cfg, exists := sv.symbols[key]
	if !exists {
		return nil, &SymbolError{
			Message: "交易对未找到",
			Symbol:  symbol,
		}
	}

	return cfg, nil
}

// RegisterSymbol 向验证器注册新的交易对
func (sv *SymbolValidator) RegisterSymbol(cfg *SymbolConfig) error {
	symbol := BuildSymbol(cfg.BaseAsset, cfg.QuoteAsset)
	sv.symbols[symbol] = cfg
	return nil
}

// FormatSymbolInfo 返回格式化的交易对信息字符串
func FormatSymbolInfo(base, quote string) string {
	var builder strings.Builder
	builder.WriteString("交易对: ")
	builder.WriteString(strings.ToUpper(base))
	builder.WriteString("/")
	builder.WriteString(strings.ToUpper(quote))
	builder.WriteString(" (")
	builder.WriteString(BuildSymbol(base, quote))
	builder.WriteString(")")
	return builder.String()
}

// SymbolError symbol解析或验证时的错误
type SymbolError struct {
	Message string // 错误信息
	Symbol  string // 涉及的symbol（可选）
}

func (e *SymbolError) Error() string {
	if e.Symbol != "" {
		return e.Symbol + ": " + e.Message
	}
	return e.Message
}

// ValidateSymbolParams 根据交易对配置验证订单参数
func (sv *SymbolValidator) ValidateSymbolParams(symbol string, price, quantity float64) error {
	cfg, err := sv.GetSymbolConfig(symbol)
	if err != nil {
		return err
	}

	// 校验数量是否在最小/最大范围内
	if quantity < cfg.MinQty {
		return &SymbolError{
			Message: "数量低于最小值",
			Symbol:  symbol,
		}
	}
	if quantity > cfg.MaxQty {
		return &SymbolError{
			Message: "数量超过最大值",
			Symbol:  symbol,
		}
	}

	// 校验成交额
	notional := price * quantity
	if notional < cfg.MinNotional {
		return &SymbolError{
			Message: "成交额低于最小值",
			Symbol:  symbol,
		}
	}

	return nil
}
