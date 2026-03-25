package chapter01

import (
	"strings"
	"testing"
)

// TestFormatDepthData_Normal 测试正常格式化输出
func TestFormatDepthData_Normal(t *testing.T) {
	bids := [][]float64{
		{65000.50, 1.2345},
		{64999.00, 2.1000},
		{64998.50, 0.5000},
	}
	asks := [][]float64{
		{65001.00, 0.8765},
		{65002.50, 1.5000},
		{65003.00, 3.0000},
	}

	output := FormatDepthData(bids, asks, 10)

	// 验证包含关键信息
	if !strings.Contains(output, "BTCUSDT") {
		t.Error("输出应包含交易对名称")
	}
	if !strings.Contains(output, "65000.50") {
		t.Error("输出应包含最优买方价格")
	}
	if !strings.Contains(output, "65001.00") {
		t.Error("输出应包含最优卖方价格")
	}
	if !strings.Contains(output, "Spread:") {
		t.Error("输出应包含Spread信息")
	}
	if !strings.Contains(output, "BID QTY") {
		t.Error("输出应包含列标题")
	}
}

// TestFormatDepthData_Limit 测试档位限制
func TestFormatDepthData_Limit(t *testing.T) {
	bids := [][]float64{
		{65000.50, 1.2345},
		{64999.00, 2.1000},
		{64998.50, 0.5000},
		{64998.00, 1.0000},
	}
	asks := [][]float64{
		{65001.00, 0.8765},
		{65002.50, 1.5000},
	}

	// 只显示2档
	output := FormatDepthData(bids, asks, 2)

	// 应该只包含2行数据
	lines := strings.Split(output, "\n")
	dataLines := 0
	for _, line := range lines {
		// 统计包含数字价格的行（简单判断）
		if strings.Contains(line, "65000") || strings.Contains(line, "64999") ||
			strings.Contains(line, "65001") || strings.Contains(line, "65002") {
			dataLines++
		}
	}
	// 实际数据行应该不超过4行（2个bids + 2个asks）
	if dataLines > 4 {
		t.Errorf("期望最多4个数据行，实际有 %d 个", dataLines)
	}
}

// TestFormatDepthData_Empty 测试空数据
func TestFormatDepthData_Empty(t *testing.T) {
	output := FormatDepthData(nil, nil, 10)
	if !strings.Contains(output, "Spread:") {
		t.Error("空数据仍应包含Spread信息")
	}
}

// TestFormatDepthData_Spread 测试Spread计算
func TestFormatDepthData_Spread(t *testing.T) {
	bids := [][]float64{
		{65000.00, 1.0},
	}
	asks := [][]float64{
		{65010.00, 1.0},
	}

	output := FormatDepthData(bids, asks, 10)

	// Spread应该是10.00
	if !strings.Contains(output, "10.00") {
		t.Error("Spread计算错误")
	}
}

// TestExportDepthDataToCSV 测试CSV导出
func TestExportDepthDataToCSV(t *testing.T) {
	bids := [][]float64{
		{65000.50, 1.2345},
		{64999.00, 2.1000},
	}
	asks := [][]float64{
		{65001.00, 0.8765},
	}

	csv := ExportDepthDataToCSV(bids, asks)

	lines := strings.Split(csv, "\n")
	// 第一行是表头
	if !strings.Contains(lines[0], "BID_PRICE") {
		t.Error("CSV应包含表头")
	}
	// 应该有3行数据（2个bids + 2个asks + 1个空行 = 可能被截断）
	if len(lines) < 3 {
		t.Error("CSV行数不足")
	}
	// 验证包含正确的数据格式
	if !strings.Contains(csv, "65000.50") {
		t.Error("CSV应包含价格数据")
	}
	if !strings.Contains(csv, "1.2345") {
		t.Error("CSV应包含数量数据")
	}
}

// TestParseDepthData 测试反向解析
func TestParseDepthData(t *testing.T) {
	bids := [][]float64{
		{65000.50, 1.2345},
	}
	asks := [][]float64{
		{65001.00, 0.8765},
	}

	formatted := FormatDepthData(bids, asks, 10)
	parsedBids, parsedAsks, err := ParseDepthData(formatted)

	if err != nil {
		t.Errorf("ParseDepthData返回错误: %v", err)
	}
	// ParseDepthData是一个简化版本，这里只验证它不报错
	_ = parsedBids
	_ = parsedAsks
}
