package chapter01

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// makeTestRequest 创建一个用于测试的Gin上下文和响应记录器
func makeTestRequest(symbol, side, orderType, price, quantity, clientOID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/order", nil)
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if symbol != "" {
		c.Request.PostForm["symbol"] = []string{symbol}
	}
	if side != "" {
		c.Request.PostForm["side"] = []string{side}
	}
	if orderType != "" {
		c.Request.PostForm["type"] = []string{orderType}
	}
	if price != "" {
		c.Request.PostForm["price"] = []string{price}
	}
	if quantity != "" {
		c.Request.PostForm["quantity"] = []string{quantity}
	}
	if clientOID != "" {
		c.Request.PostForm["clientOID"] = []string{clientOID}
	}

	return c, w
}

// TestParseOrderParams_NormalOrders 正常订单的测试用例
func TestParseOrderParams_NormalOrders(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		side        string
		orderType   string
		price       string
		quantity    string
		clientOID   string
		wantSymbol  string
		wantSide    string
		wantType    string
		wantPrice   float64
		wantQty     float64
		wantClient  string
		wantErr     bool
		errContains string
	}{
		{
			name:        "正常限价买单",
			symbol:      "BTCUSDT",
			side:        "BUY",
			orderType:   "LIMIT",
			price:       "50000.00",
			quantity:    "0.1234",
			clientOID:   "order-001",
			wantSymbol:  "BTCUSDT",
			wantSide:    "BUY",
			wantType:    "LIMIT",
			wantPrice:   50000.00,
			wantQty:     0.1234,
			wantClient:  "order-001",
			wantErr:     false,
		},
		{
			name:        "正常市价卖单",
			symbol:      "ETHUSDT",
			side:        "SELL",
			orderType:   "MARKET",
			price:       "",
			quantity:    "1.5000",
			clientOID:   "",
			wantSymbol:  "ETHUSDT",
			wantSide:    "SELL",
			wantType:    "MARKET",
			wantPrice:   0,
			wantQty:     1.5000,
			wantClient:  "",
			wantErr:     false,
		},
		{
			name:        "小写side不区分大小写",
			symbol:      "bnbusdt",
			side:        "buy",
			orderType:   "limit",
			price:       "300.50",
			quantity:    "10.0",
			clientOID:   "",
			wantSymbol:  "BNBUSDT",
			wantSide:    "BUY",
			wantType:    "LIMIT",
			wantPrice:   300.50,
			wantQty:     10.0,
			wantErr:     false,
		},
		{
			name:        "混合大小写side",
			symbol:      "ADAUSDT",
			side:        "SeLl",
			orderType:   "LIMIT",
			price:       "0.55",
			quantity:    "100",
			clientOID:   "",
			wantSymbol:  "ADAUSDT",
			wantSide:    "SELL",
			wantType:    "LIMIT",
			wantPrice:   0.55,
			wantQty:     100,
			wantErr:     false,
		},
		{
			name:        "价格额外精度被规范化",
			symbol:      "DOGEUSDT",
			side:        "BUY",
			orderType:   "LIMIT",
			price:       "0.12345",
			quantity:    "1000",
			clientOID:   "",
			wantSymbol:  "DOGEUSDT",
			wantSide:    "BUY",
			wantType:    "LIMIT",
			wantPrice:   0.12, // 规范化为2位小数
			wantQty:     1000,
			wantErr:     false,
		},
		{
			name:        "数量超过4位小数",
			symbol:      "SOLUSDT",
			side:        "BUY",
			orderType:   "LIMIT",
			price:       "100",
			quantity:    "1.123456",
			clientOID:   "",
			wantErr:     true,
			errContains: "小数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := makeTestRequest(tt.symbol, tt.side, tt.orderType, tt.price, tt.quantity, tt.clientOID)
			params, err := ParseOrderParams(c)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望有错误，实际为nil")
				}
				if tt.errContains != "" {
					orderErr, ok := err.(*OrderValidationError)
					if !ok {
						t.Fatalf("期望OrderValidationError，实际为%T", err)
					}
					if orderErr.Field == "" || !contains(orderErr.Message, tt.errContains) {
						t.Fatalf("错误%q不包含%q", orderErr.Message, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("未预期的错误: %v", err)
			}
			if params.Symbol != tt.wantSymbol {
				t.Errorf("Symbol = %v, 期望 %v", params.Symbol, tt.wantSymbol)
			}
			if params.Side != tt.wantSide {
				t.Errorf("Side = %v, 期望 %v", params.Side, tt.wantSide)
			}
			if params.Type != tt.wantType {
				t.Errorf("Type = %v, 期望 %v", params.Type, tt.wantType)
			}
			if params.Price != tt.wantPrice {
				t.Errorf("Price = %v, 期望 %v", params.Price, tt.wantPrice)
			}
			if params.Quantity != tt.wantQty {
				t.Errorf("Quantity = %v, 期望 %v", params.Quantity, tt.wantQty)
			}
			if params.ClientOID != tt.wantClient {
				t.Errorf("ClientOID = %v, 期望 %v", params.ClientOID, tt.wantClient)
			}
		})
	}
}

// TestParseOrderParams_InvalidInputs 非法输入的测试用例
func TestParseOrderParams_InvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		symbol      string
		side        string
		orderType   string
		price       string
		quantity    string
		clientOID   string
		errContains string
	}{
		{name: "空symbol", symbol: "", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "1", clientOID: "", errContains: "symbol"},
		{name: "非数字价格", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "abc", quantity: "1", clientOID: "", errContains: "数字"},
		{name: "价格为零", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "0", quantity: "1", clientOID: "", errContains: "大于0"},
		{name: "价格为负数", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "-100", quantity: "1", clientOID: "", errContains: "大于0"},
		{name: "数量为零", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "0", clientOID: "", errContains: "大于0"},
		{name: "数量为负数", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "-5", clientOID: "", errContains: "大于0"},
		{name: "非法side", symbol: "BTCUSDT", side: "HOLD", orderType: "LIMIT", price: "50000", quantity: "1", clientOID: "", errContains: "side"},
		{name: "非法type", symbol: "BTCUSDT", side: "BUY", orderType: "FOK", price: "50000", quantity: "1", clientOID: "", errContains: "type"},
		{name: "clientOID过长", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "1", clientOID: "this-is-a-very-long-client-order-id-that-exceeds-the-maximum-length-of-64-characters", errContains: "64"},
		{name: "symbol过短", symbol: "BTC", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "1", clientOID: "", errContains: "4"},
		{name: "symbol包含特殊字符", symbol: "BTC-USDT", side: "BUY", orderType: "LIMIT", price: "50000", quantity: "1", clientOID: "", errContains: "字母和数字"},
		{name: "NaN检查", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "12.34.56", quantity: "1", clientOID: "", errContains: "数字"},
		{name: "价格超过2位小数", symbol: "BTCUSDT", side: "BUY", orderType: "LIMIT", price: "50000.123", quantity: "1", clientOID: "", errContains: "小数"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := makeTestRequest(tt.symbol, tt.side, tt.orderType, tt.price, tt.quantity, tt.clientOID)
			_, err := ParseOrderParams(c)
			if err == nil {
				t.Fatalf("期望有错误，实际为nil")
			}
			orderErr, ok := err.(*OrderValidationError)
			if !ok {
				t.Fatalf("期望OrderValidationError，实际为%T: %v", err, err)
			}
			fieldMatch := orderErr.Field == "" || contains(tt.errContains, orderErr.Field) || contains(orderErr.Message, tt.errContains)
			if !fieldMatch {
				t.Errorf("错误 = {Field: %q, Message: %q}, 期望包含%q", orderErr.Field, orderErr.Message, tt.errContains)
			}
		})
	}
}

// TestIsValidPrecision 精度校验测试
func TestIsValidPrecision(t *testing.T) {
	tests := []struct {
		value       float64
		maxDecimals int
		want        bool
	}{
		{123.45, 2, true},
		{123.4, 2, true},
		{123, 2, true},
		{123.456, 2, false},
		{123.1234, 4, true},
		{123.12345, 4, false},
		{0.1, 1, true},
		{0.01, 1, false},
		{0.0001, 4, true},
		{0.00001, 4, false},
		{99.99, 2, true},
		{0.123456789, 8, true},
		{0.123456789, 9, true},
	}

	for _, tt := range tests {
		got := isValidPrecision(tt.value, tt.maxDecimals)
		if got != tt.want {
			t.Errorf("isValidPrecision(%v, %d) = %v, 期望 %v", tt.value, tt.maxDecimals, got, tt.want)
		}
	}
}

// TestNormalizePrice 价格规范化测试
func TestNormalizePrice(t *testing.T) {
	tests := []struct {
		price float64
		want  string
	}{
		{50000.0, "50000.00"},
		{50000.123, "50000.12"},
		{0.5, "0.50"},
		{100, "100.00"},
		{0.01, "0.01"},
	}

	for _, tt := range tests {
		got := NormalizePrice(tt.price)
		if got != tt.want {
			t.Errorf("NormalizePrice(%v) = %v, 期望 %v", tt.price, got, tt.want)
		}
	}
}

// TestNormalizeQuantity 数量规范化测试
func TestNormalizeQuantity(t *testing.T) {
	tests := []struct {
		qty  float64
		want string
	}{
		{0.1234, "0.1234"},
		{0.1, "0.1000"},
		{100, "100.0000"},
		{0.0001, "0.0001"},
		{1.5, "1.5000"},
	}

	for _, tt := range tests {
		got := NormalizeQuantity(tt.qty)
		if got != tt.want {
			t.Errorf("NormalizeQuantity(%v) = %v, 期望 %v", tt.qty, got, tt.want)
		}
	}
}

// TestOrderValidationError 错误类型测试
func TestOrderValidationError(t *testing.T) {
	err := &OrderValidationError{Field: "price", Message: "must be positive"}
	if err.Error() != "price: must be positive" {
		t.Errorf("OrderValidationError.Error() = %q, 期望 %q", err.Error(), "price: must be positive")
	}
}

// contains 检查字符串s是否包含substr
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
