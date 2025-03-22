package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/itherunder/go-coin-sol/type"
)

type SwapDataType struct {
	TokenAddress                 solana.PublicKey `json:"token_address"`
	SOLAmountWithDecimals        uint64           `json:"sol_amount_with_decimals"`
	TokenAmountWithDecimals      uint64           `json:"token_amount_with_decimals"`
	Type                         type_.SwapType   `json:"type"`
	UserAddress                  solana.PublicKey `json:"user_address"`
	UserTokenBalanceWithDecimals uint64           `json:"user_token_balance_with_decimals"`
}
