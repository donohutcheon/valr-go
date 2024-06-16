package streaming

import (
	"github.com/shopspring/decimal"
)

type order struct {
	ID     string          `json:"id"`
	Price  decimal.Decimal `json:"price,string"`
	Volume decimal.Decimal `json:"volume,string"`
}

type orderBook struct {
	Sequence int64    `json:"sequence,string"`
	Asks     []*order `json:"asks"`
	Bids     []*order `json:"bids"`
	Status   string   `json:"status"`
}

type TradeUpdate struct {
	// Sequence is the monotonically increasing sequence number of the trade
	// for the market in question.
	Sequence int64 `json:"sequence"`
	// Base is the volume of the base currency that was filled.
	Base decimal.Decimal `json:"base,string"`
	// Counter is the price at which the order filled.
	Counter decimal.Decimal `json:"counter,string"`
	// MakerOrderID is the ID of the pre-existing order in the order book that was matched.
	MakerOrderID string `json:"maker_order_id"`
	// TakeOrderID is the ID of the order that matched against a pre-existing order.
	TakerOrderID string `json:"taker_order_id"`
	// Deprecated: Use MakerOrderID.
	OrderID string `json:"order_id"`
	// IsBuy indicates whether the taker matched an order in the ask order book.
	IsBuy bool `json:"is_buy"`
}

type CreateUpdate struct {
	OrderID string          `json:"order_id"`
	Type    string          `json:"type"`
	Price   decimal.Decimal `json:"price,string"`
	Volume  decimal.Decimal `json:"volume,string"`
}

type DeleteUpdate struct {
	OrderID string `json:"order_id"`
}

type StatusUpdate struct {
	Status string `json:"status"`
}

type Update struct {
	Sequence     int64          `json:"sequence,string"`
	TradeUpdates []*TradeUpdate `json:"trade_updates"`
	CreateUpdate *CreateUpdate  `json:"create_update"`
	DeleteUpdate *DeleteUpdate  `json:"delete_update"`
	StatusUpdate *StatusUpdate  `json:"status_update"`
	Timestamp    int64          `json:"timestamp"`
}

type credentials struct {
	APIKeyID     string `json:"api_key_id"`
	APIKeySecret string `json:"api_key_secret"`
}
