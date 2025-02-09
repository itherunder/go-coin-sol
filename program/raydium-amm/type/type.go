package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type RaydiumSwapKeys struct {
	AmmAddress                   solana.PublicKey
	AmmOpenOrdersAddress         solana.PublicKey
	AmmTargetOrdersAddress       solana.PublicKey
	PoolCoinTokenAccountAddress  solana.PublicKey
	PoolPcTokenAccountAddress    solana.PublicKey
	SerumProgramAddress          solana.PublicKey
	SerumMarketAddress           solana.PublicKey
	SerumBidsAddress             solana.PublicKey
	SerumAsksAddress             solana.PublicKey
	SerumEventQueueAddress       solana.PublicKey
	SerumCoinVaultAccountAddress solana.PublicKey
	SerumPcVaultAccountAddress   solana.PublicKey
	SerumVaultSignerAddress      solana.PublicKey
}

type SwapDataType struct {
	TokenAddress                       solana.PublicKey `json:"token_address"`
	SOLAmountWithDecimals              uint64           `json:"sol_amount_with_decimals"`
	TokenAmountWithDecimals            uint64           `json:"token_amount_with_decimals"`
	Type                               type_.SwapType   `json:"type"`
	UserAddress                        solana.PublicKey `json:"user_address"`
	UserBalanceWithDecimals            uint64           `json:"user_balance_with_decimals"`
	BeforeUserBalanceWithDecimals      uint64           `json:"before_user_balance_with_decimals"`
	BeforeUserTokenBalanceWithDecimals uint64           `json:"before_user_token_balance_with_decimals"`
	UserTokenBalanceWithDecimals       uint64           `json:"user_token_balance_with_decimals"`
	RaydiumSwapKeys                    RaydiumSwapKeys  `json:"raydium_swap_keys,omitempty"`
	CoinIsSOL                          bool             `json:"coin_is_sol"`
}

type SwapTxDataType struct {
	Swaps   []*SwapDataType
	FeeInfo *type_.FeeInfo
	TxId    string
}
