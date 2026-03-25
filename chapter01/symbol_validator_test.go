package chapter01

import (
	"strings"
	"testing"
)

// TestParseSymbol 测试ParseSymbol函数的各种输入格式
func TestParseSymbol(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		wantBase    string
		wantQuote   string
		wantErr     bool
		errContains string
	}{
		// 无分隔符格式
		{name: "无分隔符BTCUSDT", symbol: "BTCUSDT", wantBase: "BTC", wantQuote: "USDT", wantErr: false},
		{name: "无分隔符ETHUSDT", symbol: "ETHUSDT", wantBase: "ETH", wantQuote: "USDT", wantErr: false},
		{name: "无分隔符BTCBTC", symbol: "BTCBTC", wantBase: "BTC", wantQuote: "BTC", wantErr: false},
		{name: "无分隔符ETHBTC", symbol: "ETHBTC", wantBase: "ETH", wantQuote: "BTC", wantErr: false},

		// 斜杠分隔符格式
		{name: "斜杠分隔BTC/USDT", symbol: "BTC/USDT", wantBase: "BTC", wantQuote: "USDT", wantErr: false},
		{name: "斜杠分隔ETH/BTC", symbol: "ETH/BTC", wantBase: "ETH", wantQuote: "BTC", wantErr: false},
		{name: "斜杠分隔小写", symbol: "btc/usdt", wantBase: "BTC", wantQuote: "USDT", wantErr: false},

		// 横杠分隔符格式
		{name: "横杠分隔BTC-USDT", symbol: "BTC-USDT", wantBase: "BTC", wantQuote: "USDT", wantErr: false},
		{name: "横杠分隔SOL-USDT", symbol: "SOL-USDT", wantBase: "SOL", wantQuote: "USDT", wantErr: false},
		{name: "横杠分隔混合大小写", symbol: "Eth-Usdt", wantBase: "ETH", wantQuote: "USDT", wantErr: false},

		// 空白处理
		{name: "首尾空白", symbol: "  BTCUSDT  ", wantBase: "BTC", wantQuote: "USDT", wantErr: false},
		{name: "分隔符周围空白", symbol: "BTC / USDT", wantBase: "BTC", wantQuote: "USDT", wantErr: false},
		{name: "横杠周围空白", symbol: "ETH - BTC", wantBase: "ETH", wantQuote: "BTC", wantErr: false},

		// 错误用例
		{name: "空字符串", symbol: "", wantErr: true, errContains: "空"},
		{name: "计价资产不在白名单", symbol: "BTCUSD", wantErr: true, errContains: "白名单"},
		{name: "基础资产是计价资产", symbol: "USDTUSDT", wantErr: true, errContains: "基础资产不能是计价资产"},
		{name: "仅有计价资产", symbol: "USDT", wantErr: true, errContains: "基础资产不能为空"},
		{name: "仅有斜杠前缀", symbol: "/USDT", wantErr: true, errContains: "基础资产不能为空"},
		{name: "仅有斜杠后缀", symbol: "BTC/", wantErr: true, errContains: "计价资产不能为空"},
		{name: "非法计价资产", symbol: "BTC/ABC", wantErr: true, errContains: "白名单"},
		{name: "基础资产含数字", symbol: "BTC1USDT", wantErr: true, errContains: "只能包含字母和数字"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, quote, err := ParseSymbol(tt.symbol)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望有错误包含%q，实际为nil", tt.errContains)
				}
				symErr, ok := err.(*SymbolError)
				if !ok {
					t.Fatalf("期望SymbolError，实际为%T: %v", err, err)
				}
				if !strings.Contains(symErr.Error(), tt.errContains) {
					t.Errorf("错误 %q 不包含 %q", symErr.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("未预期的错误: %v", err)
			}
			if base != tt.wantBase {
				t.Errorf("base = %q, 期望 %q", base, tt.wantBase)
			}
			if quote != tt.wantQuote {
				t.Errorf("quote = %q, 期望 %q", quote, tt.wantQuote)
			}
		})
	}
}

// TestBuildSymbol 测试BuildSymbol函数的拼接功能
func TestBuildSymbol(t *testing.T) {
	tests := []struct {
		name  string
		base  string
		quote string
		want  string
	}{
		{"大写", "BTC", "USDT", "BTCUSDT"},
		{"小写", "btc", "usdt", "BTCUSDT"},
		{"混合大小写", "Btc", "Usdt", "BTCUSDT"},
		{"首尾空白", " BTC ", " USDT ", "BTCUSDT"},
		{"ETHBTC", "ETH", "BTC", "ETHBTC"},
		{"SOLBUSD", "SOL", "BUSD", "SOLBUSD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSymbol(tt.base, tt.quote)
			if got != tt.want {
				t.Errorf("BuildSymbol(%q, %q) = %q, 期望 %q", tt.base, tt.quote, got, tt.want)
			}
		})
	}
}

// TestSymbolValidator_GetSymbolConfig 测试SymbolValidator获取配置
func TestSymbolValidator_GetSymbolConfig(t *testing.T) {
	sv := NewSymbolValidator()
	tests := []struct {
		name      string
		symbol    string
		wantBase  string
		wantQuote string
		wantErr   bool
	}{
		{"已注册BTCUSDT", "BTCUSDT", "BTC", "USDT", false},
		{"已注册ETHUSDT", "ETHUSDT", "ETH", "USDT", false},
		{"已注册ETHBTC", "ETHBTC", "ETH", "BTC", false},
		{"斜杠格式", "BTC/USDT", "BTC", "USDT", false},
		{"横杠格式", "ETH-BTC", "ETH", "BTC", false},
		{"未注册交易对", "XRPUSDT", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := sv.GetSymbolConfig(tt.symbol)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望有错误，实际为nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("未预期的错误: %v", err)
			}
			if cfg.BaseAsset != tt.wantBase || cfg.QuoteAsset != tt.wantQuote {
				t.Errorf("得到 (%s, %s), 期望 (%s, %s)", cfg.BaseAsset, cfg.QuoteAsset, tt.wantBase, tt.wantQuote)
			}
		})
	}
}

// TestSymbolValidator_ValidateSymbolParams 测试订单参数校验
func TestSymbolValidator_ValidateSymbolParams(t *testing.T) {
	sv := NewSymbolValidator()
	tests := []struct {
		name      string
		symbol    string
		price     float64
		quantity  float64
		wantErr   bool
		errSubstr string
	}{
		{name: "有效BTCUSDT订单", symbol: "BTCUSDT", price: 50000.0, quantity: 0.1, wantErr: false},
		{name: "数量低于最小值", symbol: "BTCUSDT", price: 50000.0, quantity: 0.00001, wantErr: true, errSubstr: "最低"},
		{name: "成交额低于最小值", symbol: "BTCUSDT", price: 100.0, quantity: 0.0001, wantErr: true, errSubstr: "成交额"},
		{name: "DOGE整数数量", symbol: "DOGEUSDT", price: 0.1, quantity: 100, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.ValidateSymbolParams(tt.symbol, tt.price, tt.quantity)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望有错误，实际为nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("错误 %q 不包含 %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("未预期的错误: %v", err)
			}
		})
	}
}

// TestFormatSymbolInfo 测试格式化的交易对信息输出
func TestFormatSymbolInfo(t *testing.T) {
	info := FormatSymbolInfo("BTC", "USDT")
	expected := "交易对: BTC/USDT (BTCUSDT)"
	if info != expected {
		t.Errorf("FormatSymbolInfo = %q, 期望 %q", info, expected)
	}
}
