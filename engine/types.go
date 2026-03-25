package engine

// OrderSide represents the direction of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeLimit   OrderType = "LIMIT"
	OrderTypeMarket  OrderType = "MARKET"
	OrderTypeStop    OrderType = "STOP"
)

// Order represents a trading order in the matching engine.
type Order struct {
	Symbol    string
	Side      OrderSide
	Type      OrderType
	Price     float64
	Quantity  float64
	ClientOID string
}

// OrderBook represents the order book for a trading pair.
type OrderBook struct {
	Symbol string
	Bids   [][]float64 // [[price, quantity], ...]
	Asks   [][]float64 // [[price, quantity], ...]
}

// Trade represents a matched trade between a buyer and seller.
type Trade struct {
	ID         uint32
	Price      float64
	Quantity   float64
	BuyerID    uint32
	SellerID   uint32
	Timestamp  int64
	SymbolHash uint32
	Side       OrderSide
}
