package type_

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_amm_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pkg/errors"
)

type SwapKeys struct {
	AmmAddress solana.PublicKey
	CoinMint   solana.PublicKey
	PCMint     solana.PublicKey
	Vaults     map[solana.PublicKey]solana.PublicKey // mint -> vault
}

func (t *SwapKeys) ToAccounts(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	inputToken solana.PublicKey,
	outputToken solana.PublicKey,
) ([]*solana.AccountMeta, error) {
	userInputAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		inputToken,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, inputToken)
	}

	userOutputAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		outputToken,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, outputToken)
	}

	return []*solana.AccountMeta{
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.AmmAddress,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  raydium_amm_constant.Raydium_Authority_V4[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},

		{
			PublicKey:  t.Vaults[t.CoinMint],
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  t.Vaults[t.PCMint],
			IsSigner:   false,
			IsWritable: true,
		},

		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  userInputAssociatedAccount,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  userOutputAssociatedAccount,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  userAddress,
			IsSigner:   true,
			IsWritable: false,
		},
	}, nil
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

	SwapKeys

	PairAddress solana.PublicKey `json:"pair_address"`
	SOLVault    solana.PublicKey `json:"sol_vault"`
	TokenVault  solana.PublicKey `json:"token_vault"`

	FeeInfo *type_.FeeInfo `json:"fee_info"`
}
