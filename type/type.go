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

type FeeInfo struct {
	BaseFee          string `json:"base_fee"`
	PriorityFee      string `json:"priority_fee"`
	TotalFee         string `json:"total_fee"`
	ComputeUnitPrice uint64 `json:"compute_unit_price"`
}
