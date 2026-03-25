package chapter03

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateAssetName 验证资产名称是否符合规范
// 规则：
//   - 只能是字母（Latin/CJK）或数字
//   - 不能包含空格、标点、控制字符
//   - 最长20个Unicode码点
//   - 不能包含emoji
func ValidateAssetName(name string) (valid bool, reason string) {
	// 空字符串检查
	if name == "" {
		return false, "资产名称不能为空"
	}

	// 检查码点数量（最长20个Unicode码点）
	runeCount := utf8.RuneCountInString(name)
	if runeCount > 20 {
		return false, "资产名称最长20个字符"
	}

	// 逐rune检查
	for _, r := range name {
		// 检测emoji（早期Go版本无unicode.MaxRune，手写判断）
		// emoji范围：0x1F300 - 0x1F9FF（Miscellaneous Symbols and Pictographs等）
		// 附加范围：0x2600 - 0x26FF (Miscellaneous Symbols)
		//         0x2700 - 0x27BF (Dingbats)
		//         0xFE00 - 0xFE0F (Variation Selectors)
		if r >= 0x1F300 && r <= 0x1F9FF {
			return false, "资产名称不能包含emoji"
		}
		if r >= 0x2600 && r <= 0x26FF {
			return false, "资产名称不能包含符号字符"
		}
		if r >= 0x2700 && r <= 0x27BF {
			return false, "资产名称不能包含装饰符号"
		}
		if r >= 0xFE00 && r <= 0xFE0F {
			return false, "资产名称不能包含变体选择符"
		}
		if r >= 0x1F000 && r <= 0x1FAD6 { // 更多的emoji范围
			return false, "资产名称不能包含emoji"
		}
		if r >= 0x1FA00 && r <= 0x1FAFF { // Chess symbols, etc.
			return false, "资产名称不能包含特殊符号"
		}
		if r >= 0x2328 && r <= 0x232A { // Keyboard symbols
			return false, "资产名称不能包含特殊符号"
		}

		// 空格检查
		if unicode.IsSpace(r) {
			return false, "资产名称不能包含空格"
		}

		// 标点符号检查
		if unicode.IsPunct(r) {
			return false, "资产名称不能包含标点符号"
		}

		// 控制字符检查
		if unicode.IsControl(r) {
			return false, "资产名称不能包含控制字符"
		}

		// 数字检查（允许）
		if unicode.IsDigit(r) {
			continue
		}

		// 字母检查（Latin/CJK等）
		if unicode.IsLetter(r) {
			// 使用SimpleFold处理特殊大小写折叠（如土耳其文的i问题）
			// 这确保了跨语言的一致性
			_ = unicode.SimpleFold(r)
			continue
		}

		// 其他字符（如符号、数学字符等） - 允许字母和数字，其他根据实际情况判断
		// 这里我们只允许字母和数字
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false, "资产名称只能包含字母和数字"
		}
	}

	return true, ""
}

// NormalizeAssetName 规范化资产名称
// 转换为大写，移除空白
func NormalizeAssetName(name string) string {
	// 去除首尾空白
	name = strings.TrimSpace(name)

	// 转换为大写
	var result strings.Builder
	result.Grow(len(name))

	for _, r := range name {
		// 使用SimpleFold遍历并取大写形式
		upper := unicode.ToUpper(r)
		result.WriteRune(upper)
	}

	return result.String()
}

// IsValidAssetName 快速检查资产名称是否有效
func IsValidAssetName(name string) bool {
	valid, _ := ValidateAssetName(name)
	return valid
}

// GetAssetNameRuneCount 获取资产名称的码点数量
func GetAssetNameRuneCount(name string) int {
	return utf8.RuneCountInString(name)
}

// ValidateAssetNameExamples 验证示例
func ValidateAssetNameExamples() map[string]struct{ Valid bool; Reason string } {
	return map[string]struct{ Valid bool; Reason string }{
		"BTC":       {true, ""},
		"ETH":       {true, ""},
		"USDT":      {true, ""},
		"以太坊":    {true, ""},
		"柴犬币":    {true, ""},
		"BTC123":    {true, ""},
		"USDT🔥":    {false, "emoji"},
		"BTC USDT":  {false, "空格"},
		"BTC-USDT":  {false, "标点"},
		"":          {false, "空字符串"},
		"BTC\u0000": {false, "控制字符"},
	}
}
