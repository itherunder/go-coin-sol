package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type RaydiumSwapKeys struct {
	AmmAddress                   solana.PublicKey
	AmmOpenOrdersAddress         *solana.PublicKey
	AmmTargetOrdersAddress       *solana.PublicKey
	PoolCoinTokenAccountAddress  solana.PublicKey
	PoolPcTokenAccountAddress    solana.PublicKey
	SerumProgramAddress          *solana.PublicKey
	SerumMarketAddress           *solana.PublicKey
	SerumBidsAddress             *solana.PublicKey
	SerumAsksAddress             *solana.PublicKey
	SerumEventQueueAddress       *solana.PublicKey
	SerumCoinVaultAccountAddress *solana.PublicKey
	SerumPcVaultAccountAddress   *solana.PublicKey
	SerumVaultSignerAddress      *solana.PublicKey
}

type SwapDataType struct {
	TokenAddress     solana.PublicKey `json:"token_address"`
	SOLAmount        string           `json:"sol_amount"`
	TokenAmount      string           `json:"token_amount"`
	Type             type_.SwapType   `json:"type"`
	UserAddress      solana.PublicKey `json:"user_address"`
	UserBalance      string           `json:"user_balance"`
	UserTokenBalance string           `json:"user_token_balance"`
	RaydiumSwapKeys  RaydiumSwapKeys  `json:"raydium_swap_keys,omitempty"`
}

type SwapTxDataType struct {
	Swaps   []*SwapDataType
	FeeInfo *type_.FeeInfo
	TxId    string
}
