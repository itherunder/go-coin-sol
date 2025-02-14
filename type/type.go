package type_

import (
	"github.com/gagliardetto/solana-go"
)

type DexPlatform string

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

type SwapDataType struct {
	InputAddress             solana.PublicKey `json:"input_address"`
	OutputAddress            solana.PublicKey `json:"output_address"`
	InputAmountWithDecimals  uint64           `json:"input_amount_with_decimals"`
	OutputAmountWithDecimals uint64           `json:"output_amount_with_decimals"`
	InputDecimals            uint64           `json:"input_decimals"`
	OutputDecimals           uint64           `json:"output_decimals"`
	UserAddress              solana.PublicKey `json:"user_address"`

	PairAddress solana.PublicKey `json:"pair_address"`
	InputVault  solana.PublicKey `json:"input_vault"`
	OutputVault solana.PublicKey `json:"output_vault"`

	ParsedKeys interface{} `json:"parsed_keys"`
	ExtraDatas interface{} `json:"extra_datas"`

	Program  solana.PublicKey   `json:"program"`
	Keys     []solana.PublicKey `json:"keys"`
	MethodId string             `json:"method_id"`
}

type SOLTradeInfoType struct {
	TokenAddress            solana.PublicKey `json:"token_address"`
	Type                    SwapType         `json:"type"`
	SOLAmountWithDecimals   uint64           `json:"sol_amount_with_decimals"`
	TokenAmountWithDecimals uint64           `json:"token_amount_with_decimals"`
	UserAddress             solana.PublicKey `json:"user_address"`
}

func (t *SwapDataType) ToSOLTradeInfo() *SOLTradeInfoType {
	if t.InputAddress.Equals(solana.SolMint) {
		return &SOLTradeInfoType{
			TokenAddress:            t.OutputAddress,
			Type:                    SwapType_Buy,
			SOLAmountWithDecimals:   t.InputAmountWithDecimals,
			TokenAmountWithDecimals: t.OutputAmountWithDecimals,
			UserAddress:             t.UserAddress,
		}
	} else if t.OutputAddress.Equals(solana.SolMint) {
		return &SOLTradeInfoType{
			TokenAddress:            t.InputAddress,
			Type:                    SwapType_Sell,
			SOLAmountWithDecimals:   t.OutputAmountWithDecimals,
			TokenAmountWithDecimals: t.InputAmountWithDecimals,
			UserAddress:             t.UserAddress,
		}
	} else {
		return nil
	}
}

type SwapTxDataType struct {
	Swaps   []*SwapDataType
	FeeInfo *FeeInfo
	TxId    string
}
