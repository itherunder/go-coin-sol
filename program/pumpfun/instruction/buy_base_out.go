package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	constant "github.com/pefish/go-coin-sol/constant"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	"github.com/pkg/errors"
)

type BuyInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewBuyBaseOutInstruction(
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	bondingCurveAddress solana.PublicKey,
	userAssociatedTokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	maxSolAmountWithDecimals uint64,
) (*BuyInstruction, error) {
	bondingCurveAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(
		bondingCurveAddress,
		tokenAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	methodBytes, err := hex.DecodeString("66063d1201daebea")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals  uint64
		MaxSolAmountWithDecimals uint64
	}{
		TokenAmountWithDecimals:  tokenAmountWithDecimals,
		MaxSolAmountWithDecimals: maxSolAmountWithDecimals,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &BuyInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  pumpfun_constant.Global,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  pumpfun_constant.Fee_Receiver,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  tokenAddress,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  bondingCurveAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  bondingCurveAssociatedTokenAddress,
				IsSigner:   false,
				IsWritable: true,
			},

			{
				PublicKey:  userAssociatedTokenAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userAddress,
				IsSigner:   false,
				IsWritable: true,
			},

			{
				PublicKey:  solana.SystemProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  solana.TokenProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  constant.Rent,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  pumpfun_constant.Pumpfun_Event_Authority,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  pumpfun_constant.Pumpfun_Program,
				IsSigner:   false,
				IsWritable: false,
			},
		},
		data:      append(methodBytes, params.Bytes()...),
		programID: pumpfun_constant.Pumpfun_Program,
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
