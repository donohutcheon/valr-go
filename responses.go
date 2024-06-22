package valr

import (
	"time"

	"github.com/shopspring/decimal"
)

/*
PUBLIC API GET RESPONSES
*/

// GetServerTimeResponse is the struct that GetServerTime responses are unpacked into
type GetServerTimeResponse struct {
	EpochTime int       `json:"epochTime"`
	Time      time.Time `json:"time"`
}

/*
PRIVATE API GET RESPONSES
*/

// GetDepositAddressResponse is the struct that GetDepositAddress responses are unpacked into
type GetDepositAddressResponse struct {
	Currency string `json:"currency"`
	Address  string `json:"address"`
}

// GetWithdrawInfoResponse is the struct that GetWithdrawInfo responses are unpacked into
type GetWithdrawInfoResponse struct {
	Currency                 string          `json:"currency"`
	MinimumWithdrawAmount    decimal.Decimal `json:"minimumWithdrawAmount"`
	Active                   bool            `json:"isActive"`
	WithdrawCost             decimal.Decimal `json:"withdrawCost"`
	SupportsPaymentReference bool            `json:"supportsPaymentReference"`
}

// GetSimpleBuyOrSellOrderStatusResponse is the struct that GetSimpleBuyOrSellOrderStatus responses are unpacked into
type GetSimpleBuyOrSellOrderStatusResponse struct {
	OrderID         string          `json:"orderId"`
	Success         bool            `json:"success"`
	Processing      bool            `json:"processing"`
	PaidAmount      decimal.Decimal `json:"paidAmount"`
	PaidCurrency    string          `json:"paidCurrency"`
	ReceiveAmount   decimal.Decimal `json:"receivedAmount"`
	FeeAmount       decimal.Decimal `json:"feeAmount"`
	FeeCurrency     string          `json:"feeCurrency"`
	OrderExecutedAt time.Time       `json:"orderExecutedAt"`
}

// GetOrderStatusByOrderIDResponse is the struct that GetOrderStatusByOrderID responses are unpacked into
type GetOrderStatusByOrderIDResponse struct {
	OrderID           string          `json:"orderId"`
	OrderStatusType   bool            `json:"orderStatusType"`
	CurrencyPair      string          `json:"currencyPair"`
	OriginalPrice     decimal.Decimal `json:"originalPrice"`
	RemainingQuantity decimal.Decimal `json:"remainingQuantity"`
	OriginalQuantity  decimal.Decimal `json:"originalQuantity"`
	OrderSide         ResponseSide    `json:"orderSide"`
	OrderType         string          `json:"orderType"`
	FailedReason      string          `json:"failedReason"`
	CustomerOrderID   string          `json:"customerOrderId"`
	OrderUpdatedAt    time.Time       `json:"orderUpdatedAt"`
	OrderCreatedAt    time.Time       `json:"orderCreatedAt"`
}

// GetOrderHistorySummaryByOrderIDResponse is the struct that GetOrderHistorySummaryByOrderID responses are unpacked into
type GetOrderHistorySummaryByOrderIDResponse struct {
	OrderID           string          `json:"orderId"`
	OrderStatusType   bool            `json:"orderStatusType"`
	Pair              string          `json:"currencyPair"`
	AveragePrice      decimal.Decimal `json:"averagePrice"`
	OriginalPrice     decimal.Decimal `json:"originalPrice"`
	RemainingQuantity decimal.Decimal `json:"remainingQuantity"`
	OriginalQuantity  decimal.Decimal `json:"originalQuantity"`
	Total             decimal.Decimal `json:"total"`
	TotalFee          decimal.Decimal `json:"totalFee"`
	FeeCurrency       string          `json:"feeCurrency"`
	OrderSide         ResponseSide    `json:"orderSide"`
	OrderType         string          `json:"orderType"`
	FailedReason      string          `json:"failedReason"`
	OrderUpdatedAt    time.Time       `json:"orderUpdatedAt"`
	OrderCreatedAt    time.Time       `json:"orderCreatedAt"`
}

// GetOrderHistorySummaryByCustomerOrderIDResponse is the struct that GetOrderHistorySummaryByCustomerOrderID responses are unpacked into
type GetOrderHistorySummaryByCustomerOrderIDResponse struct {
	OrderID           string          `json:"orderId"`
	CustomerOrderID   string          `json:"customerOrderId"`
	OrderStatusType   bool            `json:"orderStatusType"`
	Pair              string          `json:"currencyPair"`
	AveragePrice      decimal.Decimal `json:"averagePrice"`
	OriginalPrice     decimal.Decimal `json:"originalPrice"`
	RemainingQuantity decimal.Decimal `json:"remainingQuantity"`
	OriginalQuantity  decimal.Decimal `json:"originalQuantity"`
	Total             decimal.Decimal `json:"total"`
	TotalFee          decimal.Decimal `json:"totalFee"`
	FeeCurrency       string          `json:"feeCurrency"`
	OrderSide         ResponseSide    `json:"orderSide"`
	OrderType         string          `json:"orderType"`
	FailedReason      string          `json:"failedReason"`
	OrderUpdatedAt    time.Time       `json:"orderUpdatedAt"`
	OrderCreatedAt    time.Time       `json:"orderCreatedAt"`
}

/*
PRIVATE API POST RESPONSES
*/

// PostNewCryptoWithdrawResponse is the struct that PostNewCryptoWithdraw responses are unpacked into
type PostNewCryptoWithdrawResponse struct {
	ID string `json:"id"`
}

// PostNewFiatWithdrawResponse is the struct that PostNewFiatWithdraw responses are unpacked into
type PostNewFiatWithdrawResponse struct {
	ID string `json:"id"`
}

// PostSimpleBuyOrSellQuoteResponse is the struct that PostSimpleBuyOrSellQuote responses are unpacked into
type PostSimpleBuyOrSellQuoteResponse struct {
	Pair          string          `json:"currencyPair"`
	PayAmount     decimal.Decimal `json:"payAmount"`
	ReceiveAmount decimal.Decimal `json:"receiveAmount"`
	Fee           decimal.Decimal `json:"fee"`
	FeeCurrency   string          `json:"feeCurrency"`
	CreatedAt     time.Time       `json:"createdAt"`
	OrderID       string          `json:"id"`
}

// PostSimpleBuyOrSellOrderResponse is the struct that PostSimpleBuyOrSellOrder responses are unpacked into
type PostSimpleBuyOrSellOrderResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PostLimitOrderResponse is the struct that PostLimitOrder responses are unpacked into
type PostLimitOrderResponse struct {
	ID string `json:"id"`
}

// PostMarketOrderResponse is the struct that PostMarketOrder responses are unpacked into
type PostMarketOrderResponse struct {
	ID string `json:"id"`
}

/*
PRIVATE API DEL RESPONSES
*/

// DelOrderResponse is the struct that DelOrder responses are unpacked into
type DelOrderResponse struct {
	// Empty 202 Response
}

// DelOrderByCustomerOrderIDResponse is the struct that DelOrder responses are unpacked into
type DelOrderByCustomerOrderIDResponse struct {
	// Empty 202 Response
}
