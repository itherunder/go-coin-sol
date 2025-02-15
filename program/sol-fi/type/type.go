package type_

import (
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

type SwapKeys struct {
	PairAddress solana.PublicKey
	MintA       solana.PublicKey
	MintB       solana.PublicKey
	Vaults      map[solana.PublicKey]solana.PublicKey
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
			PublicKey:  t.Vaults[t.MintA],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  t.Vaults[t.MintB],
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintAAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  userMintBAssociatedAccount,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  solana.SysVarInstructionsPubkey,
			IsWritable: false,
			IsSigner:   false,
		},
	}, nil
}
