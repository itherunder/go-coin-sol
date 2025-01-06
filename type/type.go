package type_

type SwapType string

const (
	SwapType_Buy  SwapType = "buy"
	SwapType_Sell SwapType = "sell"
)

type TokenAmountInfo struct {
	AmountWithDecimals uint64 `json:"amount_with_decimals"`
	Decimals           uint64 `json:"decimals"`
}

type FeeInfo struct {
	BaseFeeWithDecimals     uint64 `json:"base_fee_with_decimals"`
	PriorityFeeWithDecimals uint64 `json:"priority_fee_with_decimals"`
	TotalFeeWithDecimals    uint64 `json:"total_fee_with_decimals"`
	ComputeUnitPrice        uint64 `json:"compute_unit_price"`
}
