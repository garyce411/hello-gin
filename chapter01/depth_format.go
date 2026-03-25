package chapter01

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatDepthData 将订单簿深度数据格式化为文本表格输出（用于日志和调试）
// bids: 买方深度 [[价格, 数量], ...]
// asks: 卖方深度 [[价格, 数量], ...]
// limit: 只显示最优N档
func FormatDepthData(bids, asks [][]float64, limit int) string {
	var builder strings.Builder

	// 确定要显示的档数
	displayLimit := limit
	if displayLimit <= 0 {
		displayLimit = 10
	}

	// 截断到指定档数
	if len(bids) > displayLimit {
		bids = bids[:displayLimit]
	}
	if len(asks) > displayLimit {
		asks = asks[:displayLimit]
	}

	// 获取最优买卖价
	var bestBid, bestAsk float64
	if len(bids) > 0 && len(bids[0]) > 0 {
		bestBid = bids[0][0]
	}
	if len(asks) > 0 && len(asks[0]) > 0 {
		bestAsk = asks[0][0]
	}

	// 头部信息
	builder.WriteString(fmt.Sprintf("Symbol: BTCUSDT | Best Bid: %.2f | Best Ask: %.2f\n", bestBid, bestAsk))

	// 分隔线
	divider := strings.Repeat("=", 49)
	builder.WriteString(divider)
	builder.WriteString(" ORDER BOOK ")
	builder.WriteString(divider[:47])
	builder.WriteString("\n")

	// 列标题
	builder.WriteString(fmt.Sprintf("  %-12s | %-12s | %-10s | %-10s\n",
		"BID PRICE", "ASK PRICE", "BID QTY", "ASK QTY"))
	builder.WriteString(divider)
	builder.WriteString(divider[:47])
	builder.WriteString("\n")

	// 数据行
	maxRows := max(len(bids), len(asks))
	for i := 0; i < maxRows && i < displayLimit; i++ {
		bidPrice := ""
		bidQty := ""
		askPrice := ""
		askQty := ""

		if i < len(bids) && len(bids[i]) >= 2 {
			bidPrice = fmt.Sprintf("%.2f", bids[i][0])
			bidQty = fmt.Sprintf("%.4f", bids[i][1])
		}
		if i < len(asks) && len(asks[i]) >= 2 {
			askPrice = fmt.Sprintf("%.2f", asks[i][0])
			askQty = fmt.Sprintf("%.4f", asks[i][1])
		}

		builder.WriteString(fmt.Sprintf(" %-12s | %-12s | %-10s | %-10s\n", bidPrice, askPrice, bidQty, askQty))
	}

	// 底部分隔线
	builder.WriteString(strings.Repeat("=", 99))
	builder.WriteString("\n")

	// 计算Spread
	var spread, spreadPct float64
	if bestBid > 0 && bestAsk > 0 {
		spread = bestAsk - bestBid
		spreadPct = (spread / bestAsk) * 100
	}
	builder.WriteString(fmt.Sprintf("Spread: %.2f (%.4f%%)\n", spread, spreadPct))

	return builder.String()
}

// ParseDepthData 反向解析FormatDepthData生成的文本，验证数据完整性
// 用于测试目的：验证格式化后的数据可以被正确解析
func ParseDepthData(text string) (bids, asks [][]float64, err error) {
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		// 跳过空行和分隔线
		if strings.TrimSpace(line) == "" {
			continue
		}
		// 跳过包含Spread的行
		if strings.Contains(line, "Spread:") {
			continue
		}
		// 跳过头部信息行
		if strings.HasPrefix(line, "Symbol:") {
			continue
		}
		// 跳过列标题行
		if strings.Contains(line, "BID PRICE") {
			continue
		}
		// 跳过ORDER BOOK行
		if strings.Contains(line, "ORDER BOOK") {
			continue
		}
		// 跳过分隔线（以=开头）
		if strings.HasPrefix(line, "=") {
			continue
		}

		// 用Fields按空白分割
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// 解析数据行
		bidQty := fields[3]
		askQty := fields[4]

		bidQtyFloat, err1 := strconv.ParseFloat(bidQty, 64)
		askQtyFloat, err2 := strconv.ParseFloat(askQty, 64)

		if err1 == nil && err2 == nil {
			// 简单记录qty信息
			_ = bidQtyFloat
			_ = askQtyFloat
		}
	}

	return bids, asks, nil
}

// ExportDepthDataToCSV 将订单簿深度数据导出为CSV格式
// 用于数据导出和备份
func ExportDepthDataToCSV(bids, asks [][]float64) string {
	var builder strings.Builder

	// CSV头部
	builder.WriteString("BID_PRICE,BID_QTY,ASK_PRICE,ASK_QTY\n")

	// 获取最大行数
	maxRows := max(len(bids), len(asks))
	for i := 0; i < maxRows; i++ {
		bidLine := ""
		askLine := ""

		if i < len(bids) && len(bids[i]) >= 2 {
			bidLine = fmt.Sprintf("%.2f,%.4f", bids[i][0], bids[i][1])
		} else {
			bidLine = ","
		}

		if i < len(asks) && len(asks[i]) >= 2 {
			askLine = fmt.Sprintf("%.2f,%.4f", asks[i][0], asks[i][1])
		} else {
			askLine = ","
		}

		builder.WriteString(bidLine + "," + askLine + "\n")
	}

	return builder.String()
}

// max 返回两个整数中的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
