package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type ExtraDatasType struct {
	ReserveSOLAmountWithDecimals   uint64 `json:"reserve_sol_amount_with_decimals"`
	ReserveTokenAmountWithDecimals uint64 `json:"reserve_token_amount_with_decimals"`
	Timestamp                      uint64 `json:"timestamp"`
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
