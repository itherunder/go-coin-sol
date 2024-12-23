package type_

type SwapType string

const (
	SwapType_Buy  SwapType = "buy"
	SwapType_Sell SwapType = "sell"
)

type TokenAmountInfo struct {
	Amount   string `json:"amount"`
	Decimals uint64 `json:"decimals"`
}
