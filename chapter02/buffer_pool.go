package chapter02

import (
	"bytes"
	"sync"
)

// OrderBufferPool 订单缓冲区池，复用bytes.Buffer以减少GC压力
// 在撮合引擎的成交记录写入路径中使用，可以显著减少内存分配
type OrderBufferPool struct {
	pool sync.Pool
	size int // buffer初始大小
}

// NewOrderBufferPool 创建一个新的缓冲区池
// size参数设置buffer的初始容量
func NewOrderBufferPool(size int) *OrderBufferPool {
	return &OrderBufferPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				// 新建buffer时设置初始容量
				buf := &bytes.Buffer{}
				if size > 0 {
					buf.Grow(size)
				}
				return buf
			},
		},
	}
}

// Get 从池中获取一个buffer，自动调用Reset()
// 调用者获得buffer后应尽快归还
func (p *OrderBufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset() // 重置buffer内容但保留容量
	return buf
}

// Put 将buffer归还到池中
// 如果buffer过大（>1MB），则不回收直接丢弃
func (p *OrderBufferPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}
	// 超大buffer不回收，避免内存泄漏
	if b.Cap() > 1<<20 { // 1MB
		return
	}
	p.pool.Put(b)
}

// WriteTradeToBuffer 使用buffer池将成交记录写入字符串
// 这是一个使用示例，演示如何在实际场景中使用buffer池
func (p *OrderBufferPool) WriteTradeToBuffer(tradeID uint32, price, quantity float64, symbol string) string {
	buf := p.Get()
	defer p.Put(buf)

	// 写入成交信息
	buf.WriteString("Trade ID: ")
	buf.WriteString(formatUint32(tradeID))
	buf.WriteString(", Symbol: ")
	buf.WriteString(symbol)
	buf.WriteString(", Price: ")
	buf.WriteString(formatFloat(price))
	buf.WriteString(", Qty: ")
	buf.WriteString(formatFloat(quantity))

	return buf.String()
}

// GetPoolStats 返回池的当前状态（用于监控）
func (p *OrderBufferPool) GetPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"initialSize": p.size,
	}
}

// formatUint32 将uint32转换为字符串（避免strconv导入）
func formatUint32(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// formatFloat 将float64格式化为字符串（简化版）
func formatFloat(f float64) string {
	// 使用简单的方式处理
	var buf bytes.Buffer
	buf.Grow(20)
	intPart := int64(f)
	fracPart := int64((f - float64(intPart)) * 10000)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	buf.WriteString(formatUint32(uint32(intPart)))
	buf.WriteByte('.')
	if fracPart < 1000 {
		buf.WriteByte('0')
	}
	if fracPart < 100 {
		buf.WriteByte('0')
	}
	if fracPart < 10 {
		buf.WriteByte('0')
	}
	buf.WriteString(formatUint32(uint32(fracPart)))
	return buf.String()
}
