package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/pefish/go-coin-sol/constant"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	raydium_type "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
	go_decimal "github.com/pefish/go-decimal"
)

type SellInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewSellBaseInInstruction(
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	userWSOLAssociatedAccount solana.PublicKey,
	userTokenAssociatedAccount solana.PublicKey,
	tokenAmount type_.TokenAmountInfo,
	minReceiveSOLAmount string,
	raydiumSwapKeys raydium_type.RaydiumSwapKeys,
) (*SellInstruction, error) {
	methodBytes, err := hex.DecodeString("09")
	if err != nil {
		return nil, err
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals uint64
		MinReceiveSOLAmount     uint64
	}{
		TokenAmountWithDecimals: go_decimal.Decimal.MustStart(tokenAmount.Amount).MustShiftedBy(tokenAmount.Decimals).RoundDown(0).MustEndForUint64(),
		MinReceiveSOLAmount:     go_decimal.Decimal.MustStart(minReceiveSOLAmount).MustShiftedBy(constant.SOL_Decimals).RoundDown(0).MustEndForUint64(),
	})
	if err != nil {
		return nil, err
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
				PublicKey:  raydium_constant.Raydium_Authority_V4,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.AmmOpenOrdersAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.AmmOpenOrdersAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.AmmTargetOrdersAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.AmmTargetOrdersAddress
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
					if raydiumSwapKeys.SerumProgramAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumProgramAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumMarketAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumMarketAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumBidsAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumBidsAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumAsksAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumAsksAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumEventQueueAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumEventQueueAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumCoinVaultAccountAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumCoinVaultAccountAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumPcVaultAccountAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumPcVaultAccountAddress
					}
				}(),
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey: func() solana.PublicKey {
					if raydiumSwapKeys.SerumVaultSignerAddress == nil {
						return solana.SolMint
					} else {
						return *raydiumSwapKeys.SerumVaultSignerAddress
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
		programID: raydium_constant.Raydium_Liquidity_Pool_V4,
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
