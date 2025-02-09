package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	raydium_type "github.com/pefish/go-coin-sol/program/raydium-amm/type"
	"github.com/pkg/errors"
)

type SellInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewSellBaseInInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	userWSOLAssociatedAccount solana.PublicKey,
	userTokenAssociatedAccount solana.PublicKey,
	tokenAmountWithDecimals uint64,
	minReceiveSOLAmountWithDecimals uint64,
	raydiumSwapKeys raydium_type.RaydiumSwapKeys,
) (*SellInstruction, error) {
	methodBytes, err := hex.DecodeString("09")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals uint64
		MinReceiveSOLAmount     uint64
	}{
		TokenAmountWithDecimals: tokenAmountWithDecimals,
		MinReceiveSOLAmount:     minReceiveSOLAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &SellInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  solana.TokenProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  raydiumSwapKeys.AmmAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  raydium_constant.Raydium_Authority_V4[network],
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.AmmOpenOrdersAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.AmmOpenOrdersAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.AmmTargetOrdersAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.AmmTargetOrdersAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},

			{
				PublicKey:  raydiumSwapKeys.PoolCoinTokenAccountAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  raydiumSwapKeys.PoolPcTokenAccountAddress,
				IsSigner:   false,
				IsWritable: true,
			},

			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumProgramAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumProgramAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumMarketAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumMarketAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumBidsAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumBidsAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumAsksAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumAsksAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumEventQueueAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumEventQueueAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumCoinVaultAccountAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumCoinVaultAccountAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumPcVaultAccountAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumPcVaultAccountAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumVaultSignerAddress.IsZero() {
						return solana.SolMint
					} else {
						return raydiumSwapKeys.SerumVaultSignerAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  userTokenAssociatedAccount,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userWSOLAssociatedAccount,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userAddress,
				IsSigner:   true,
				IsWritable: false,
			},
		},
		data:      append(methodBytes, params.Bytes()...),
		programID: raydium_constant.Raydium_Liquidity_Pool_V4[network],
	}, nil
}

func (t *SellInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *SellInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *SellInstruction) Data() ([]byte, error) {
	return t.data, nil
}
