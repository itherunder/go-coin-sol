package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
)

type SellInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewSellBaseInInstruction(
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	bondingCurveAddress solana.PublicKey,
	userAssociatedTokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	minSolReceiveAmountWithDecimals uint64,
) (*SellInstruction, error) {
	bondingCurveAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(
		bondingCurveAddress,
		tokenAddress,
	)
	if err != nil {
		return nil, err
	}
	methodBytes, err := hex.DecodeString("33e685a4017f83ad")
	if err != nil {
		return nil, err
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		TokenAmountWithDecimals         uint64
		MinSolReceiveAmountWithDecimals uint64
	}{
		TokenAmountWithDecimals:         tokenAmountWithDecimals,
		MinSolReceiveAmountWithDecimals: minSolReceiveAmountWithDecimals,
	})
	if err != nil {
		return nil, err
	}
	return &SellInstruction{
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
				PublicKey:  solana.SPLAssociatedTokenAccountProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  solana.TokenProgramID,
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

func (t *SellInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *SellInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *SellInstruction) Data() ([]byte, error) {
	return t.data, nil
}
