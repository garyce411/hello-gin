package chapter03

import (
	"errors"
	"unicode/utf8"
)

// ValidateUTF8Consistency 验证UTF-8编码一致性
// 遍历字节，使用utf8.FullRune判断当前字节位置是否有完整的rune
// 返回：是否有效、所有无效字节的位置列表、错误信息
func ValidateUTF8Consistency(data []byte) (bool, []int, error) {
	var invalidPositions []int
	var firstError error

	i := 0
	for i < len(data) {
		// 使用utf8.FullRune判断当前位置是否有完整的rune
		if !utf8.FullRune(data[i:]) {
			// 不完整的rune序列
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = errors.New("不完整的多字节序列")
			}
			// 跳过当前字节继续检查
			i++
			continue
		}

		// 尝试解码rune
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// 解码失败（无效字节）
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = utf8.ErrRuneInvalid
			}
		}

		// 检查是否是多字节序列长度正确
		expectedSize := utf8.RuneLen(r)
		if expectedSize != size && size > 0 {
			// 这实际上不应该发生，因为utf8.FullRune已经检查过了
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = errors.New("rune长度不匹配")
			}
		}

		// 移动到下一个rune
		i += size
	}

	return len(invalidPositions) == 0, invalidPositions, firstError
}

// ValidateUTF8Strict 严格验证UTF-8编码
// 不使用utf8.Valid，手写验证逻辑
func ValidateUTF8Strict(data []byte) (bool, []int, error) {
	var invalidPositions []int
	var firstError error

	for i := 0; i < len(data); {
		b := data[i]

		// 单字节字符 (0x00-0x7F)
		// ASCII字符，最高位为0
		if b <= 0x7F {
			i++
			continue
		}

		// 多字节字符，首字节的高位模式决定长度
		var expectedLen int
		var validLead bool

		if b&0xE0 == 0xC0 {
			// 2字节序列 (110xxxxx 10xxxxxx)
			expectedLen = 2
			validLead = true
		} else if b&0xF0 == 0xE0 {
			// 3字节序列 (1110xxxx 10xxxxxx 10xxxxxx)
			expectedLen = 3
			validLead = true
		} else if b&0xF8 == 0xF0 {
			// 4字节序列 (11110xxx 10xxxxxx 10xxxxxx 10xxxxxx)
			// UTF-8只支持到4字节，超过4字节是无效的
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = errors.New("过长的UTF-8编码（4字节以上）")
			}
			i++
			continue
		} else if b&0xC0 == 0x80 {
			// 后续字节 (10xxxxxx) 不应该单独出现作为首字节
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = utf8.ErrRuneStart
			}
			i++
			continue
		} else {
			// 无效的首字节
			invalidPositions = append(invalidPositions, i)
			if firstError == nil {
				firstError = utf8.ErrRuneInvalid
			}
			i++
			continue
		}

		// 检查后续字节是否完整
		if i+expectedLen-1 >= len(data) {
			// 数据不完整
			for j := i; j < len(data); j++ {
				invalidPositions = append(invalidPositions, j)
			}
			if firstError == nil {
				firstError = utf8.ErrRuneTooShort
			}
			break
		}

		// 验证后续字节（必须以10开头）
		valid := true
		for j := 1; j < expectedLen; j++ {
			if data[i+j]&0xC0 != 0x80 {
				valid = false
				// 将这个位置标记为无效
				invalidPositions = append(invalidPositions, i+j)
			}
		}

		if !valid {
			if firstError == nil {
				firstError = utf8.ErrRuneInvalid
			}
		}

		// 验证编码值是否在有效范围内
		// 2字节: 0x80-0x7FF (U+0080-U+07FF)
		// 3字节: 0x800-0xFFFF (U+0800-U+FFFF，不包括surrogate)
		// 4字节: 0x10000-0x10FFFF (U+10000-U+10FFFF)
		var codepoint int
		switch expectedLen {
		case 2:
			codepoint = int(b&0x1F)<<6 | int(data[i+1]&0x3F)
			if codepoint < 0x80 {
				invalidPositions = append(invalidPositions, i)
				if firstError == nil {
					firstError = utf8.ErrRuneInvalid
				}
			}
		case 3:
			codepoint = int(b&0x0F)<<12 | int(data[i+1]&0x3F)<<6 | int(data[i+2]&0x3F)
			// 检查是否在surrogate范围内
			if codepoint >= 0xD800 && codepoint <= 0xDFFF {
				invalidPositions = append(invalidPositions, i)
				if firstError == nil {
					firstError = utf8.ErrRuneInvalid
				}
			}
			if codepoint < 0x800 {
				invalidPositions = append(invalidPositions, i)
				if firstError == nil {
					firstError = utf8.ErrRuneInvalid
				}
			}
		case 4:
			codepoint = int(b&0x07)<<18 | int(data[i+1]&0x3F)<<12 | int(data[i+2]&0x3F)<<6 | int(data[i+3]&0x3F)
			if codepoint < 0x10000 {
				invalidPositions = append(invalidPositions, i)
				if firstError == nil {
					firstError = utf8.ErrRuneInvalid
				}
			}
			if codepoint > 0x10FFFF {
				invalidPositions = append(invalidPositions, i)
				if firstError == nil {
					firstError = utf8.ErrRuneInvalid
				}
			}
		}

		i += expectedLen
	}

	return len(invalidPositions) == 0, invalidPositions, firstError
}

// CountInvalidBytes 计算无效字节的数量
func CountInvalidBytes(data []byte) int {
	_, invalidPositions, _ := ValidateUTF8Consistency(data)
	return len(invalidPositions)
}

// FindFirstInvalidByte 找到第一个无效字节的位置
func FindFirstInvalidByte(data []byte) (int, error) {
	valid, invalidPositions, err := ValidateUTF8Consistency(data)
	if valid {
		return -1, nil
	}
	if len(invalidPositions) > 0 {
		return invalidPositions[0], err
	}
	return -1, err
}

// IsValidUTF8 快速检查是否是有效的UTF-8
func IsValidUTF8(data []byte) bool {
	valid, _, _ := ValidateUTF8Consistency(data)
	return valid
}
