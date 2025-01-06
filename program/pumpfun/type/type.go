package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type SwapDataType struct {
	TokenAddress                   solana.PublicKey `json:"token_address"`
	SOLAmountWithDecimals          uint64           `json:"sol_amount_with_decimals"`
	TokenAmountWithDecimals        uint64           `json:"token_amount_with_decimals"`
	Type                           type_.SwapType   `json:"type"`
	UserAddress                    solana.PublicKey `json:"user_address"`
	ReserveSOLAmountWithDecimals   uint64           `json:"reserve_sol_amount_with_decimals"`
	ReserveTokenAmountWithDecimals uint64           `json:"reserve_token_amount_with_decimals"`
	Timestamp                      uint64           `json:"timestamp"`
}

type SwapTxDataType struct {
	Swaps                         []*SwapDataType `json:"swaps"`
	FeeInfo                       *type_.FeeInfo  `json:"fee_info"`
	TxId                          string          `json:"tx_id"`
	UserBalanceWithDecimals       uint64          `json:"user_balance_with_decimals"`
	BeforeUserBalanceWithDecimals uint64          `json:"before_user_balance_with_decimals"`
}

type ParseTxResult struct {
	SwapTxData      *SwapTxDataType
	CreateTxData    *CreateTxDataType
	RemoveLiqTxData *RemoveLiqTxDataType
	AddLiqTxData    *AddLiqTxDataType
}

type CreateTxDataType struct {
	CreateDataType
	TxId    string         `json:"txid"`
	FeeInfo *type_.FeeInfo `json:"fee_info"`
}

type CreateDataType struct {
	Name                string           `json:"name"`
	Symbol              string           `json:"symbol"`
	URI                 string           `json:"uri"`
	UserAddress         solana.PublicKey `json:"user_address"`
	BondingCurveAddress solana.PublicKey `json:"bonding_curve_address"`
	TokenAddress        solana.PublicKey `json:"token_address"`
}

type RemoveLiqTxDataType struct {
	TxId                string           `json:"txid"`
	BondingCurveAddress solana.PublicKey `json:"bonding_curve_address"`
	TokenAddress        solana.PublicKey `json:"token_address"`
	FeeInfo             *type_.FeeInfo   `json:"fee_info"`
}

type AddLiqTxDataType struct {
	TxId                        string           `json:"txid"`
	TokenAddress                solana.PublicKey `json:"token_address"`
	InitSOLAmountWithDecimals   uint64           `json:"init_sol_amount_with_decimals"`
	InitTokenAmountWithDecimals uint64           `json:"init_token_amount_with_decimals"`
	AMMAddress                  solana.PublicKey `json:"amm_address"`
	PoolCoinTokenAccount        solana.PublicKey `json:"pool_coin_token_account"`
	PoolPcTokenAccount          solana.PublicKey `json:"pool_pc_token_account"`

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}
