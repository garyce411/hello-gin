package chapter03

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateSymbol 交易对符号的Unicode安全校验
// 扩展练习1-2，在符号校验中加入Unicode边界检查
type SymbolValidationResult struct {
	Valid         bool
	ErrorMessage  string
	HasBOM        bool // 字节顺序标记
	HasZWS        bool // 零宽空格
	HasRLO        bool // 反向文本覆盖
	HasInvalidUTF bool // 无效UTF-8
}

// ValidateSymbol 安全校验交易对符号
// 检查：
//   - 有效的UTF-8编码
//   - 不是以多字节字符开头
//   - 不含隐藏字符（零宽空格、BOM、反向文本等）
func ValidateSymbol(symbol string) *SymbolValidationResult {
	result := &SymbolValidationResult{Valid: true}

	// 空检查
	if symbol == "" {
		result.Valid = false
		result.ErrorMessage = "符号不能为空"
		return result
	}

	// 检查是否是有效的UTF-8
	if !utf8.ValidString(symbol) {
		result.Valid = false
		result.ErrorMessage = "符号包含无效的UTF-8编码"
		result.HasInvalidUTF = true
		return result
	}

	// 检查第一个字符
	r, size := utf8.DecodeRuneInString(symbol)
	if r == utf8.RuneError && size == 1 {
		result.Valid = false
		result.ErrorMessage = "符号以无效字符开头"
		return result
	}

	// 检查是否以多字节ASCII字符开头（应该以字母开头）
	if size > 1 {
		result.Valid = false
		result.ErrorMessage = "符号不能以多字节字符开头"
		return result
	}

	// 检查是否为字母
	if !unicode.IsLetter(r) {
		result.Valid = false
		result.ErrorMessage = "符号必须以字母开头"
		return result
	}

	// 逐rune检查隐藏字符
	for i, r := range symbol {
		// NULL字符
		if r == 0x0000 {
			result.Valid = false
			result.ErrorMessage = "符号包含NULL字符"
			return result
		}

		// 零宽空格 (Zero Width Space)
		if r == 0x200B {
			result.Valid = false
			result.ErrorMessage = "符号包含零宽空格"
			result.HasZWS = true
			return result
		}

		// 零宽非连接符 (Zero Width Non-Joiner)
		if r == 0x200C {
			result.Valid = false
			result.ErrorMessage = "符号包含零宽非连接符"
			result.HasZWS = true
			return result
		}

		// 零宽连接符 (Zero Width Joiner)
		if r == 0x200D {
			result.Valid = false
			result.ErrorMessage = "符号包含零宽连接符"
			result.HasZWS = true
			return result
		}

		// 字节顺序标记 (BOM)
		if r == 0xFEFF {
			result.Valid = false
			result.ErrorMessage = "符号包含BOM"
			result.HasBOM = true
			return result
		}

		// 反向文本覆盖 (Right-to-Left Override)
		if r == 0x202E {
			result.Valid = false
			result.ErrorMessage = "符号包含反向文本覆盖字符"
			result.HasRLO = true
			return result
		}

		// 左到右覆盖 (Left-to-Right Override)
		if r == 0x202D {
			result.Valid = false
			result.ErrorMessage = "符号包含左到右覆盖字符"
			return result
		}

		// 代理对（Surrogate Pair）- 这在Go字符串中不应该出现
		if r >= 0xD800 && r <= 0xDFFF {
			result.Valid = false
			result.ErrorMessage = "符号包含无效的代理对"
			return result
		}

		// 使用SimpleFold检测变体形式
		if i == 0 {
			// 对第一个字符使用SimpleFold检查
			folded := unicode.SimpleFold(r)
			if folded == r && unicode.IsUpper(r) {
				// 这是一个固定的字符（如数字或某些特殊字符）
				// 检查是否是合法的交易对字符
			}
		}

		// 跳过已检查的隐藏字符
		_ = i
	}

	// 检查是否包含土耳其文的i问题
	// 土耳其文的i有特殊的点（ı和İ），不是普通的i
	for _, r := range symbol {
		if r == 'ı' || r == 'İ' {
			result.Valid = false
			result.ErrorMessage = "符号不应包含土耳其文特殊字符"
			return result
		}
	}

	return result
}

// NormalizeSymbol 规范化交易对符号
// 移除所有非ASCII字母数字字符，用strings.Builder重建
// 逐rune处理，使用unicode.ToUpper
func NormalizeSymbol(symbol string) string {
	var builder strings.Builder
	builder.Grow(len(symbol))

	for _, r := range symbol {
		// 使用逐rune方式处理，而不是直接strings.ToUpper
		upper := unicode.ToUpper(r)

		// 只保留字母和数字
		if unicode.IsLetter(upper) || unicode.IsDigit(upper) {
			builder.WriteRune(upper)
		}
	}

	return builder.String()
}

// ToUpperRuneByRune 逐rune转换为大写
// 对比直接strings.ToUpper的差异
func ToUpperRuneByRune(symbol string) string {
	var builder strings.Builder
	builder.Grow(len(symbol))

	for _, r := range symbol {
		upper := unicode.ToUpper(r)
		builder.WriteRune(upper)
	}

	return builder.String()
}

// DetectHiddenCharacters 检测并列出符号中的隐藏字符
func DetectHiddenCharacters(symbol string) []string {
	var hidden []string

	for _, r := range symbol {
		switch r {
		case 0x0000:
			hidden = append(hidden, "NULL (U+0000)")
		case 0x200B:
			hidden = append(hidden, "零宽空格 (U+200B)")
		case 0x200C:
			hidden = append(hidden, "零宽非连接符 (U+200C)")
		case 0x200D:
			hidden = append(hidden, "零宽连接符 (U+200D)")
		case 0xFEFF:
			hidden = append(hidden, "BOM (U+FEFF)")
		case 0x202E:
			hidden = append(hidden, "反向文本覆盖 (U+202E)")
		case 0x202D:
			hidden = append(hidden, "左到右覆盖 (U+202D)")
		case 0x00A0:
			hidden = append(hidden, "不间断空格 (U+00A0)")
		case 0x1680:
			hidden = append(hidden, "Ogham空格标记 (U+1680)")
		case 0x2028:
			hidden = append(hidden, "行分隔符 (U+2028)")
		case 0x2029:
			hidden = append(hidden, "段落分隔符 (U+2029)")
		case 0x3000:
			hidden = append(hidden, "表意空格 (U+3000)")
		}
	}

	return hidden
}

// IsSafeSymbol 检查符号是否安全（不包含任何隐藏字符）
func IsSafeSymbol(symbol string) bool {
	result := ValidateSymbol(symbol)
	return result.Valid
}

// ContainsHiddenChars 检查是否包含隐藏字符
func ContainsHiddenChars(symbol string) bool {
	return len(DetectHiddenCharacters(symbol)) > 0
}
