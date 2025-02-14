package instruction

import (
	"bytes"
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
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
		return nil, errors.Wrap(err, "")
	}
	methodBytes, err := hex.DecodeString("33e685a4017f83ad")
	if err != nil {
		return nil, errors.Wrap(err, "")
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
		return nil, errors.Wrap(err, "")
	}
	return &SellInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  pumpfun_constant.Global,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  pumpfun_constant.Fee_Receiver[network],
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
				PublicKey:  pumpfun_constant.Pumpfun_Program[network],
				IsSigner:   false,
				IsWritable: false,
			},
		},
		data:      append(methodBytes, params.Bytes()...),
		programID: pumpfun_constant.Pumpfun_Program[network],
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
