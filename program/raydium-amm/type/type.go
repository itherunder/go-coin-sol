package type_

import (
	"github.com/gagliardetto/solana-go"
	type_ "github.com/pefish/go-coin-sol/type"
)

type SwapKeys struct {
	AmmAddress                  solana.PublicKey
	PoolCoinTokenAccountAddress solana.PublicKey
	PoolPcTokenAccountAddress   solana.PublicKey
	CoinMint                    solana.PublicKey
	PCMint                      solana.PublicKey
	Vaults                      map[solana.PublicKey]solana.PublicKey
}

type ExtraDatasType struct {
	ReserveInputWithDecimals  uint64 `json:"reserve_input_with_decimals"`
	ReserveOutputWithDecimals uint64 `json:"reserve_output_with_decimals"`
}

type AddLiqTxDataType struct {
	TxId                        string           `json:"txid"`
	TokenAddress                solana.PublicKey `json:"token_address"`
	InitSOLAmountWithDecimals   uint64           `json:"init_sol_amount_with_decimals"`
	InitTokenAmountWithDecimals uint64           `json:"init_token_amount_with_decimals"`

	AMMAddress           solana.PublicKey `json:"amm_address"`
	PoolCoinTokenAccount solana.PublicKey `json:"pool_coin_token_account"`
	PoolPcTokenAccount   solana.PublicKey `json:"pool_pc_token_account"`

	PairAddress solana.PublicKey `json:"pair_address"`
	SOLVault    solana.PublicKey `json:"sol_vault"`
	TokenVault  solana.PublicKey `json:"token_vault"`

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}
