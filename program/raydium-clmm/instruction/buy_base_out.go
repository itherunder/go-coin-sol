package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium-clmm/constant"
	raydium_clmm_type "github.com/pefish/go-coin-sol/program/raydium-clmm/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

type BuyInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewBuyBaseOutInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	userWSOLAssociatedAccount solana.PublicKey,
	userTokenAssociatedAccount solana.PublicKey,
	tokenAmountWithDecimals uint64,
	maxCostSolAmountWithDecimals uint64,
	swapKeys raydium_clmm_type.SwapKeys,
) (*BuyInstruction, error) {
	methodBytes, err := hex.DecodeString(util.GetDiscriminator("global", "swap_v2"))
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		Amount               uint64
		OtherAmountThreshold uint64
		SqrtPriceLimitX64    bin.Uint128
		IsBaseInput          bool
	}{
		Amount:               tokenAmountWithDecimals,
		OtherAmountThreshold: maxCostSolAmountWithDecimals,
		SqrtPriceLimitX64:    *bin.NewUint128BigEndian(),
		IsBaseInput:          false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	accounts := []*solana.AccountMeta{
		{
			PublicKey:  userAddress,
			IsSigner:   true,
			IsWritable: true,
		},
		{
			PublicKey:  raydium_constant.AMM_Config[network],
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  swapKeys.PoolIdAddress,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  userWSOLAssociatedAccount,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  userTokenAssociatedAccount,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  swapKeys.WSOLVault,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  swapKeys.TokenVault,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  swapKeys.ObservationState,
			IsSigner:   false,
			IsWritable: true,
		},
		{
			PublicKey:  solana.TokenProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.Token2022ProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.MemoProgramID,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  solana.SolMint,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  tokenAddress,
			IsSigner:   false,
			IsWritable: false,
		},
		{
			PublicKey:  swapKeys.ExBitmapAccount,
			IsSigner:   false,
			IsWritable: true,
		},
	}
	for _, remainAccount := range swapKeys.RemainAccounts {
		accounts = append(accounts, &solana.AccountMeta{
			PublicKey:  remainAccount,
			IsSigner:   false,
			IsWritable: true,
		})
	}
	return &BuyInstruction{
		accounts:  accounts,
		data:      append(methodBytes, params.Bytes()...),
		programID: raydium_constant.Raydium_Concentrated_Liquidity[network],
	}, nil
}

func (t *BuyInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *BuyInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *BuyInstruction) Data() ([]byte, error) {
	return t.data, nil
}
