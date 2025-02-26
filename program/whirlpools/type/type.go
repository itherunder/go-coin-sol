package type_

import (
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

type SwapV2Keys struct {
	PairAddress solana.PublicKey
	MintA       solana.PublicKey
	MintB       solana.PublicKey
	TickArray0  solana.PublicKey
	TickArray1  solana.PublicKey
	TickArray2  solana.PublicKey
	Oracle      solana.PublicKey

	Vaults map[solana.PublicKey]solana.PublicKey
}

func (t *SwapV2Keys) ToAccounts(
	userAddress solana.PublicKey,
) ([]*solana.AccountMeta, error) {
	userMintAAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.MintA,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.MintA)
	}

	userMintBAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.MintB,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.MintB)
	}

	return []*solana.AccountMeta{
		{
			PublicKey:  solana.TokenProgramID,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  solana.MemoProgramID,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  userAddress,
			IsWritable: true,
			IsSigner:   true,
		},
		{
			PublicKey:  t.PairAddress,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.MintA,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  t.MintB,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintAAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Vaults[t.MintA],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintBAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Vaults[t.MintB],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray0,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray1,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray2,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Oracle,
			IsWritable: true,
			IsSigner:   false,
		},
	}, nil
}

type SwapKeys struct {
	PairAddress solana.PublicKey
	TickArray0  solana.PublicKey
	TickArray1  solana.PublicKey
	TickArray2  solana.PublicKey
	Oracle      solana.PublicKey

	MintA  solana.PublicKey
	MintB  solana.PublicKey
	Vaults map[solana.PublicKey]solana.PublicKey
}

func (t *SwapKeys) ToAccounts(
	userAddress solana.PublicKey,
) ([]*solana.AccountMeta, error) {
	userMintAAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.MintA,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.MintA)
	}

	userMintBAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		t.MintB,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, t.MintB)
	}

	return []*solana.AccountMeta{
		{
			PublicKey:  solana.TokenProgramID,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  userAddress,
			IsWritable: true,
			IsSigner:   true,
		},
		{
			PublicKey:  t.PairAddress,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintAAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Vaults[t.MintA],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintBAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Vaults[t.MintB],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray0,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray1,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.TickArray2,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Oracle,
			IsWritable: true,
			IsSigner:   false,
		},
	}, nil
}

type ExtraDatasType struct {
	ReserveInputWithDecimals  uint64 `json:"reserve_input_with_decimals"`
	ReserveOutputWithDecimals uint64 `json:"reserve_output_with_decimals"`
}
