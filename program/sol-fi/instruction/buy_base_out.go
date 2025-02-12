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
	raydiumSwapKeys raydium_type.SwapKeys,
) (*BuyInstruction, error) {
	methodBytes, err := hex.DecodeString("0b")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		MaxCostSolAmountWithDecimals uint64
		TokenAmountWithDecimals      uint64
	}{
		MaxCostSolAmountWithDecimals: maxCostSolAmountWithDecimals,
		TokenAmountWithDecimals:      tokenAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &BuyInstruction{
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
				PublicKey:  userAddress,
				IsSigner:   true,
				IsWritable: false,
			},
		},
		data:      append(methodBytes, params.Bytes()...),
		programID: raydium_constant.Raydium_Liquidity_Pool_V4[network],
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
