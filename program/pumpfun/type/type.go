package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type SwapDataType struct {
	TokenAddress       solana.PublicKey `json:"token_address"`
	SOLAmount          string           `json:"sol_amount"`
	TokenAmount        string           `json:"token_amount"`
	Type               type_.SwapType   `json:"type"`
	UserAddress        solana.PublicKey `json:"user_address"`
	ReserveSOLAmount   string           `json:"reserve_sol_amount"`
	ReserveTokenAmount string           `json:"reserve_token_amount"`
	Timestamp          uint64           `json:"timestamp"`
}

type SwapTxDataType struct {
	Swaps             []*SwapDataType `json:"swaps"`
	FeeInfo           *type_.FeeInfo  `json:"fee_info"`
	TxId              string          `json:"tx_id"`
	UserBalance       string          `json:"user_balance"`
	BeforeUserBalance string          `json:"before_user_balance"`
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
	TxId                 string           `json:"txid"`
	TokenAddress         solana.PublicKey `json:"token_address"`
	InitSOLAmount        string           `json:"init_sol_amount"`
	InitTokenAmount      string           `json:"init_token_amount"`
	AMMAddress           solana.PublicKey `json:"amm_address"`
	PoolCoinTokenAccount solana.PublicKey `json:"pool_coin_token_account"`
	PoolPcTokenAccount   solana.PublicKey `json:"pool_pc_token_account"`

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}
