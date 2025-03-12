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

type CreateInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewCreateInstruction(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	tokenAddress solana.PublicKey,
	bondingCurveAddress solana.PublicKey,
	name string,
	symbol string,
	uri string,
) (*SellInstruction, error) {
	bondingCurveAssociatedTokenAddress, _, err := solana.FindAssociatedTokenAddress(
		bondingCurveAddress,
		tokenAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	methodBytes, err := hex.DecodeString("181ec828051c0777")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	params := new(bytes.Buffer)
	err = bin.NewBorshEncoder(params).Encode(struct {
		Name    string           `json:"name"`
		Symbol  string           `json:"symbol"`
		URI     string           `json:"uri"`
		Creator solana.PublicKey `json:"creator"`
	}{
		Name:    name,
		Symbol:  symbol,
		URI:     uri,
		Creator: userAddress,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	metadataAssociatedAddress, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("metadata"),
			solana.TokenMetadataProgramID.Bytes(),
			tokenAddress.Bytes(),
		},
		solana.TokenMetadataProgramID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return &SellInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  tokenAddress,
				IsSigner:   true,
				IsWritable: true,
			},
			{
				PublicKey:  pumpfun_constant.Pumpfun_Token_Mint_Authority,
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
				PublicKey:  pumpfun_constant.Global,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  solana.TokenMetadataProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  metadataAssociatedAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  userAddress,
				IsSigner:   true,
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
				PublicKey:  solana.SPLAssociatedTokenAccountProgramID,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  solana.SysVarRentPubkey,
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

func (t *CreateInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *CreateInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *CreateInstruction) Data() ([]byte, error) {
	return t.data, nil
}
