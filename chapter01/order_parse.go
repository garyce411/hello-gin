package chapter01

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OrderParams 解析HTTP请求中的下单参数
type OrderParams struct {
	Symbol    string  // 交易对，如 "BTCUSDT"
	Side      string  // "BUY" 或 "SELL"
	Type      string  // "LIMIT" 或 "MARKET"
	Price     float64 // 限价单价格（精确到0.01）
	Quantity  float64 // 数量（精确到0.0001）
	ClientOID string  // 客户端订单ID（可选，最大64字符）
}

// ParseOrderParams 从HTTP请求中解析下单参数
// 支持JSON和form表单两种格式
func ParseOrderParams(c *gin.Context) (*OrderParams, error) {
	contentType := c.ContentType()

	var symbol, side, orderType, clientOID, priceStr, quantityStr string

	// 根据Content-Type选择解析方式
	if strings.HasPrefix(contentType, "application/json") {
		// JSON格式：优先从请求头获取，fallback到query参数
		symbol = c.GetHeader("X-Symbol")
		if symbol == "" {
			symbol = c.Query("symbol")
		}
		side = c.GetHeader("X-Side")
		if side == "" {
			side = c.Query("side")
		}
		orderType = c.GetHeader("X-Type")
		if orderType == "" {
			orderType = c.Query("type")
		}
		clientOID = c.GetHeader("X-ClientOID")
		if clientOID == "" {
			clientOID = c.Query("clientOID")
		}
		priceStr = c.GetHeader("X-Price")
		if priceStr == "" {
			priceStr = c.Query("price")
		}
		quantityStr = c.GetHeader("X-Quantity")
		if quantityStr == "" {
			quantityStr = c.Query("quantity")
		}
	} else {
		// 表单格式
		symbol = c.PostForm("symbol")
		side = c.PostForm("side")
		orderType = c.PostForm("type")
		clientOID = c.PostForm("clientOID")
		priceStr = c.PostForm("price")
		quantityStr = c.PostForm("quantity")
	}

	// 去除首尾空白
	symbol = strings.TrimSpace(symbol)
	side = strings.TrimSpace(side)
	orderType = strings.TrimSpace(orderType)
	clientOID = strings.TrimSpace(clientOID)
	priceStr = strings.TrimSpace(priceStr)
	quantityStr = strings.TrimSpace(quantityStr)

	// 校验Symbol：不能为空
	if symbol == "" {
		return nil, &OrderValidationError{Field: "symbol", Message: "symbol不能为空"}
	}

	// 规范化Symbol：转换为大写
	symbol = strings.ToUpper(symbol)

	// 校验Symbol格式：只能包含字母和数字
	for _, r := range symbol {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return nil, &OrderValidationError{Field: "symbol", Message: "symbol只能包含字母和数字"}
		}
	}
	if len(symbol) < 4 {
		return nil, &OrderValidationError{Field: "symbol", Message: "symbol至少需要4个字符（BASE + QUOTE）"}
	}

	// 校验Side：必须为BUY或SELL（不区分大小写）
	sideUpper := strings.ToUpper(side)
	if sideUpper != "BUY" && sideUpper != "SELL" {
		return nil, &OrderValidationError{Field: "side", Message: "side必须为'BUY'或'SELL'"}
	}
	side = sideUpper

	// 校验Type：必须为LIMIT或MARKET（不区分大小写）
	typeUpper := strings.ToUpper(orderType)
	if typeUpper != "LIMIT" && typeUpper != "MARKET" {
		return nil, &OrderValidationError{Field: "type", Message: "type必须为'LIMIT'或'MARKET'"}
	}
	orderType = typeUpper

	// 校验ClientOID长度
	if len(clientOID) > 64 {
		return nil, &OrderValidationError{Field: "clientOID", Message: "clientOID最大64个字符"}
	}

	// 解析Price（仅限价单必需）
	var price float64
	if orderType == "LIMIT" || priceStr != "" {
		if priceStr == "" {
			return nil, &OrderValidationError{Field: "price", Message: "限价单必须指定价格"}
		}
		var err error
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, &OrderValidationError{Field: "price", Message: "无效的价格格式"}
		}
		// 检查NaN（strconv.ParseFloat对非数字输入返回NaN，如"abc"）
		if math.IsNaN(price) {
			return nil, &OrderValidationError{Field: "price", Message: "价格必须是有效数字"}
		}
		// 校验价格必须大于0
		if price <= 0 {
			return nil, &OrderValidationError{Field: "price", Message: "价格必须大于0"}
		}
		// 校验价格精度：最多2位小数
		if !isValidPrecision(price, 2) {
			return nil, &OrderValidationError{Field: "price", Message: "价格最多2位小数"}
		}
	}

	// 解析Quantity
	var quantity float64
	if quantityStr == "" {
		return nil, &OrderValidationError{Field: "quantity", Message: "数量不能为空"}
	}
	var err error
	quantity, err = strconv.ParseFloat(quantityStr, 64)
	if err != nil {
		return nil, &OrderValidationError{Field: "quantity", Message: "无效的数量格式"}
	}
	// 检查NaN
	if math.IsNaN(quantity) {
		return nil, &OrderValidationError{Field: "quantity", Message: "数量必须是有效数字"}
	}
	// 校验数量必须大于0
	if quantity <= 0 {
		return nil, &OrderValidationError{Field: "quantity", Message: "数量必须大于0"}
	}
	// 校验数量精度：最多4位小数
	if !isValidPrecision(quantity, 4) {
		return nil, &OrderValidationError{Field: "quantity", Message: "数量最多4位小数"}
	}

	// 规范化价格到固定精度
	if price > 0 {
		price, _ = strconv.ParseFloat(strconv.FormatFloat(price, 'f', 2, 64), 64)
	}

	return &OrderParams{
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		Price:     price,
		Quantity:  quantity,
		ClientOID: clientOID,
	}, nil
}

// isValidPrecision 检查float64值是否有最多指定位数的小数
// 通过将值乘以10^n，然后检查结果是否接近整数来实现
func isValidPrecision(value float64, maxDecimals int) bool {
	multiplier := math.Pow(10, float64(maxDecimals))
	scaled := value * multiplier
	diff := scaled - math.Floor(scaled)
	return diff < 0.0001 || diff > 0.9999
}

// NormalizePrice 将价格格式化为精确2位小数
func NormalizePrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}

// NormalizeQuantity 将数量格式化为精确4位小数
func NormalizeQuantity(quantity float64) string {
	return strconv.FormatFloat(quantity, 'f', 4, 64)
}

// OrderValidationError 订单参数校验错误
type OrderValidationError struct {
	Field   string // 错误字段名
	Message string // 错误信息
}

func (e *OrderValidationError) Error() string {
	return e.Field + ": " + e.Message
}
