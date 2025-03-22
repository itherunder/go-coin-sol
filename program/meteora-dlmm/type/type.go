package type_

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	meteora_dlmm_constant "github.com/pefish/go-coin-sol/program/meteora-dlmm/constant"
	"github.com/pkg/errors"
)

type SwapKeys struct {
	PairAddress    solana.PublicKey
	Oracle         solana.PublicKey
	RemainAccounts []solana.PublicKey
	MintX          solana.PublicKey
	MintY          solana.PublicKey
	Vaults         map[solana.PublicKey]solana.PublicKey // mint -> vault
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

	accounts := []*solana.AccountMeta{
		{
			PublicKey:  t.PairAddress,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  meteora_dlmm_constant.Meteora_DLMM_Program[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.Vaults[t.MintX],
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  t.Vaults[t.MintY],
			IsSigner:   false,
			IsWritable: true,
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
			PublicKey:  t.MintX,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.MintY,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  t.Oracle,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  meteora_dlmm_constant.Meteora_DLMM_Program[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  userAddress,
			IsSigner:   true,
			IsWritable: true,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  meteora_dlmm_constant.Meteora_DLMM_Event_Authority[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  meteora_dlmm_constant.Meteora_DLMM_Program[network],
			IsSigner:   false,
			IsWritable: false,
		},
	}

	for _, remainAccount := range t.RemainAccounts {
		accounts = append(accounts, &solana.AccountMeta{
			PublicKey:  remainAccount,
			IsSigner:   false,
			IsWritable: true,
		})
	}

	return accounts, nil
}

type ExtraDatasType struct {
	ReserveInputWithDecimals  uint64 `json:"reserve_input_with_decimals"`
	ReserveOutputWithDecimals uint64 `json:"reserve_output_with_decimals"`
}
