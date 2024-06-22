package streaming

import "time"

type MessageType struct {
	Type string `json:"type"`
}

type MessageTradeUpdate struct {
	MessageType
	CurrencyPairSymbol string `json:"currencyPairSymbol"`
	Data               struct {
		Price        string    `json:"price"`
		Quantity     string    `json:"quantity"`
		CurrencyPair string    `json:"currencyPair"`
		TradedAt     time.Time `json:"tradedAt"`
		TakerSide    string    `json:"takerSide"`
		ID           string    `json:"id"`
	} `json:"data"`
}

type Subscriptions struct {
	Event string   `json:"event"`
	Pairs []string `json:"pairs"`
}

type SubscribeToMarketsRequest struct {
	Type          string          `json:"type"`
	Subscriptions []Subscriptions `json:"subscriptions"`
}
