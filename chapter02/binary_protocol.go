package chapter02

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Trade 成交记录结构体
// 对应48字节的二进制协议格式
type Trade struct {
	ID         uint32 // 4字节: 成交ID
	Price      uint64 // 8字节: 价格（精度1e-8）
	Quantity   uint64 // 8字节: 数量（精度1e-8）
	BuyerID    uint32 // 4字节: 买方用户ID
	SellerID   uint32 // 4字节: 卖方用户ID
	Timestamp  int64  // 8字节: 时间戳（Unix毫秒）
	SymbolHash uint32 // 4字节: 交易对哈希
	Side       uint8  // 1字节: 方向 (1=BUY, 2=SELL)
	_          [3]byte // 3字节: 填充对齐
	_          uint32 // 4字节: 保留字段
}

// TradeProtocolSize 成交记录的固定大小（48字节）
const TradeProtocolSize = 48

// SerializeTrade 将成交记录序列化为字节数组
// 返回固定48字节的数据
func SerializeTrade(t *Trade) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("trade is nil")
	}

	// 使用固定大小的buffer
	buf := new(bytes.Buffer)
	buf.Grow(TradeProtocolSize)

	// 按协议格式写入各字段（大端序）
	if err := binary.Write(buf, binary.BigEndian, t.ID); err != nil {
		return nil, fmt.Errorf("failed to write ID: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.Price); err != nil {
		return nil, fmt.Errorf("failed to write Price: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.Quantity); err != nil {
		return nil, fmt.Errorf("failed to write Quantity: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.BuyerID); err != nil {
		return nil, fmt.Errorf("failed to write BuyerID: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.SellerID); err != nil {
		return nil, fmt.Errorf("failed to write SellerID: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.Timestamp); err != nil {
		return nil, fmt.Errorf("failed to write Timestamp: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, t.SymbolHash); err != nil {
		return nil, fmt.Errorf("failed to write SymbolHash: %w", err)
	}
	// 写入Side字段
	if err := binary.Write(buf, binary.BigEndian, t.Side); err != nil {
		return nil, fmt.Errorf("failed to write Side: %w", err)
	}
	// 写入3字节填充
	padding := [3]byte{0, 0, 0}
	if err := binary.Write(buf, binary.BigEndian, padding); err != nil {
		return nil, fmt.Errorf("failed to write padding: %w", err)
	}
	// 写入保留字段
	reserved := uint32(0)
	if err := binary.Write(buf, binary.BigEndian, reserved); err != nil {
		return nil, fmt.Errorf("failed to write reserved: %w", err)
	}

	// 验证长度
	data := buf.Bytes()
	if len(data) != TradeProtocolSize {
		return nil, fmt.Errorf("expected %d bytes, got %d", TradeProtocolSize, len(data))
	}

	return data, nil
}

// DeserializeTrade 从字节数组反序列化成交记录
// data必须至少48字节
func DeserializeTrade(data []byte) (*Trade, error) {
	if len(data) < TradeProtocolSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", TradeProtocolSize, len(data))
	}

	// 使用bytes.Reader读取
	reader := bytes.NewReader(data[:TradeProtocolSize])
	bufReader := bufio.NewReader(reader)

	var trade Trade
	var err error

	// 按协议格式读取各字段（大端序）
	trade.ID, err = readUint32(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID: %w", err)
	}
	trade.Price, err = readUint64(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Price: %w", err)
	}
	trade.Quantity, err = readUint64(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Quantity: %w", err)
	}
	trade.BuyerID, err = readUint32(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read BuyerID: %w", err)
	}
	trade.SellerID, err = readUint32(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read SellerID: %w", err)
	}
	trade.Timestamp, err = readInt64(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Timestamp: %w", err)
	}
	trade.SymbolHash, err = readUint32(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read SymbolHash: %w", err)
	}
	trade.Side, err = readUint8(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Side: %w", err)
	}
	// 跳过3字节填充
	if err := skipBytes(bufReader, 3); err != nil {
		return nil, fmt.Errorf("failed to skip padding: %w", err)
	}
	// 读取保留字段
	_, err = readUint32(bufReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read reserved: %w", err)
	}

	return &trade, nil
}

// SerializeTrades 序列化多个成交记录
func SerializeTrades(trades []*Trade) ([]byte, error) {
	var result []byte
	for _, t := range trades {
		data, err := SerializeTrade(t)
		if err != nil {
			return nil, err
		}
		result = append(result, data...)
	}
	return result, nil
}

// DeserializeTrades 反序列化多个成交记录
func DeserializeTrades(data []byte) ([]*Trade, error) {
	var trades []*Trade
	offset := 0

	for offset+TradeProtocolSize <= len(data) {
		trade, err := DeserializeTrade(data[offset : offset+TradeProtocolSize])
		if err != nil {
			return nil, fmt.Errorf("at offset %d: %w", offset, err)
		}
		trades = append(trades, trade)
		offset += TradeProtocolSize
	}

	// 如果还有剩余字节（不足以构成完整记录）
	if offset != len(data) {
		return nil, fmt.Errorf("incomplete trade data: %d bytes remaining", len(data)-offset)
	}

	return trades, nil
}

// CompareTrades 比较两笔成交是否完全一致
// 使用bytes.Equal进行字节级别的比较
func CompareTrades(t1, t2 *Trade) bool {
	if t1 == nil || t2 == nil {
		return t1 == t2
	}

	data1, err1 := SerializeTrade(t1)
	data2, err2 := SerializeTrade(t2)
	if err1 != nil || err2 != nil {
		return false
	}

	return bytes.Equal(data1, data2)
}

// ParseBatchTrades 使用bytes.SplitN解析批次二进制记录
// 返回每条记录的起始偏移量
func ParseBatchTrades(data []byte) ([][]byte, error) {
	if len(data)%TradeProtocolSize != 0 {
		return nil, fmt.Errorf("data length %d is not a multiple of %d", len(data), TradeProtocolSize)
	}

	var records [][]byte
	for i := 0; i < len(data); i += TradeProtocolSize {
		records = append(records, data[i:i+TradeProtocolSize])
	}
	return records, nil
}

// SplitTradesBySide 使用bytes.SplitN按方向分割成交记录
func SplitTradesBySide(data []byte) (buys, sells [][]byte, err error) {
	records, err := ParseBatchTrades(data)
	if err != nil {
		return nil, nil, err
	}

	for _, record := range records {
		// Side字段在偏移量40处（ID(4) + Price(8) + Quantity(8) + BuyerID(4) + SellerID(4) + Timestamp(8) + SymbolHash(4) = 40）
		side := record[40]
		if side == 1 { // BUY
			buys = append(buys, record)
		} else { // SELL
			sells = append(sells, record)
		}
	}

	return buys, sells, nil
}

// ===== 辅助函数 =====

func readUint32(r *bufio.Reader) (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

func readUint64(r *bufio.Reader) (uint64, error) {
	var v uint64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

func readInt64(r *bufio.Reader) (int64, error) {
	var v int64
	err := binary.Read(r, binary.BigEndian, &v)
	return v, err
}

func readUint8(r *bufio.Reader) (uint8, error) {
	b, err := r.ReadByte()
	return uint8(b), err
}

func skipBytes(r *bufio.Reader, n int) error {
	_, err := r.Discard(n)
	return err
}

// PriceFromUint64 将uint64格式的价格转换为float64
// 精度为1e-8
func PriceFromUint64(price uint64) float64 {
	return float64(price) / 1e8
}

// PriceToUint64 将float64格式的价格转换为uint64
// 精度为1e-8
func PriceToUint64(price float64) uint64 {
	return uint64(price * 1e8)
}

// QuantityFromUint64 将uint64格式的数量转换为float64
// 精度为1e-8
func QuantityFromUint64(qty uint64) float64 {
	return float64(qty) / 1e8
}

// QuantityToUint64 将float64格式的数量转换为uint64
// 精度为1e-8
func QuantityToUint64(qty float64) uint64 {
	return uint64(qty * 1e8)
}

// FormatTradeAsText 将成交记录格式化为可读文本
func FormatTradeAsText(t *Trade) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Trade #%d: ", t.ID))
	buf.WriteString(fmt.Sprintf("Buyer=%d, Seller=%d, ", t.BuyerID, t.SellerID))
	buf.WriteString(fmt.Sprintf("Price=%.8f, Qty=%.8f, ", PriceFromUint64(t.Price), QuantityFromUint64(t.Quantity)))
	buf.WriteString(fmt.Sprintf("Time=%d, SymbolHash=%d, Side=%s", t.Timestamp, t.SymbolHash, sideToString(t.Side)))
	return buf.String()
}

func sideToString(side uint8) string {
	switch side {
	case 1:
		return "BUY"
	case 2:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// ValidateTradeData 验证成交数据的有效性
// 使用bytes.Contains检查特定模式
func ValidateTradeData(data []byte) error {
	if len(data) < TradeProtocolSize {
		return fmt.Errorf("data too short: %d bytes", len(data))
	}

	// 检查是否有全零的ID（可能表示无效记录）
	if bytes.Equal(data[0:4], []byte{0, 0, 0, 0}) {
		return fmt.Errorf("invalid trade: zero ID")
	}

	// 检查价格是否为0（可能表示无效记录）
	if bytes.Equal(data[4:12], []byte{0, 0, 0, 0, 0, 0, 0, 0}) {
		return fmt.Errorf("invalid trade: zero price")
	}

	return nil
}

// ExtractTradeIDs 从二进制数据中提取所有成交ID
// 用于快速查找或去重
func ExtractTradeIDs(data []byte) ([]uint32, error) {
	if len(data)%TradeProtocolSize != 0 {
		return nil, fmt.Errorf("invalid data length")
	}

	var ids []uint32
	for i := 0; i < len(data); i += TradeProtocolSize {
		id := binary.BigEndian.Uint32(data[i : i+4])
		ids = append(ids, id)
	}
	return ids, nil
}

// FindTradeByID 在二进制数据中查找指定ID的成交记录
// 使用bytes.Equal进行字节比较
func FindTradeByID(data []byte, targetID uint32) (*Trade, error) {
	targetBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(targetBytes, targetID)

	for i := 0; i <= len(data)-TradeProtocolSize; i += TradeProtocolSize {
		if bytes.Equal(data[i:i+4], targetBytes) {
			return DeserializeTrade(data[i : i+TradeProtocolSize])
		}
	}

	return nil, fmt.Errorf("trade not found: %d", targetID)
}

// ===== K线数据解析 =====

// KLine K线数据结构
type KLine struct {
	Timestamp int64   // 时间戳
	Open      float64 // 开盘价
	High      float64 // 最高价
	Low       float64 // 最低价
	Close     float64 // 收盘价
	Volume    float64 // 成交量
}

// KLineParser K线数据流式解析器
// 使用bufio.Scanner进行流式处理，适合处理GB级别的文件
type KLineParser struct {
	scanner  *bufio.Scanner
	filePath string
	reader   io.ReadCloser
	lineNum  int
	parser   *KLineParserImpl
}

// KLineParserImpl K线解析的具体实现
type KLineParserImpl struct {
	pool *sync.Pool
}

// NewKLineParser 创建新的K线解析器
func NewKLineParser() *KLineParserImpl {
	return &KLineParserImpl{
		pool: &sync.Pool{
			New: func() interface{} {
				return bufio.NewReaderSize(nil, 64*1024) // 64KB buffer
			},
		},
	}
}

// ParseKLineFile 流式解析K线CSV文件
// 格式：timestamp,open,high,low,close,volume
// 不一次性读取全文件，适合GB级别数据
func (p *KLineParserImpl) ParseKLineFile(filePath string, handler func(*KLine) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 从池中获取bufio.Reader
	reader := p.pool.Get().(*bufio.Reader)
	defer p.pool.Put(reader)
	reader.Reset(file)

	// 创建scanner，配置缓冲区大小（默认64KB）
	scanner := bufio.NewScanner(reader)
	// 如果单行数据超过64KB，需要手动设置更大的缓冲区
	// scanner.Buffer(make([]byte, 100*1024), 1024*1024) // 1MB最大行

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// 跳过空行
		if len(line) == 0 {
			continue
		}

		// 跳过表头
		if lineNum == 1 && bytes.HasPrefix(line, []byte("timestamp")) {
			continue
		}

		kline, err := p.parseLine(line)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		if err := handler(kline); err != nil {
			return fmt.Errorf("handler error at line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// parseLine 解析单行K线数据
func (p *KLineParserImpl) parseLine(line []byte) (*KLine, error) {
	// 使用逗号分割
	parts := bytes.Split(line, []byte(","))
	if len(parts) < 6 {
		return nil, fmt.Errorf("expected 6 fields, got %d", len(parts))
	}

	kline := &KLine{}

	// 解析时间戳
	timestamp, err := strconv.ParseInt(string(parts[0]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	kline.Timestamp = timestamp

	// 解析各价格字段
	kline.Open, err = strconv.ParseFloat(string(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid open: %w", err)
	}
	kline.High, err = strconv.ParseFloat(string(parts[2]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid high: %w", err)
	}
	kline.Low, err = strconv.ParseFloat(string(parts[3]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid low: %w", err)
	}
	kline.Close, err = strconv.ParseFloat(string(parts[4]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid close: %w", err)
	}
	kline.Volume, err = strconv.ParseFloat(string(parts[5]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %w", err)
	}

	return kline, nil
}

// ParseKLineFromString 解析字符串格式的K线数据
// 用于测试
func (p *KLineParserImpl) ParseKLineFromString(data string) ([]*KLine, error) {
	lines := strings.Split(data, "\n")
	var klines []*KLine

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "timestamp") {
			continue
		}

		kline, err := p.parseLine([]byte(line))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		klines = append(klines, kline)
	}

	return klines, nil
}
