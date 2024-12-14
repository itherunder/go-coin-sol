package instruction

import (
	solana "github.com/gagliardetto/solana-go"
)

type CreateInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewCreateInstruction(
	payer solana.PublicKey,
	newAssociatedTokenAddress solana.PublicKey,
	addressOwner solana.PublicKey,
	tokenAddress solana.PublicKey,
) *CreateInstruction {
	return &CreateInstruction{
		accounts: []*solana.AccountMeta{
			{
				PublicKey:  payer,
				IsSigner:   true,
				IsWritable: true,
			},
			{
				PublicKey:  newAssociatedTokenAddress,
				IsSigner:   false,
				IsWritable: true,
			},
			{
				PublicKey:  addressOwner,
				IsSigner:   false,
				IsWritable: false,
			},
			{
				PublicKey:  tokenAddress,
				IsSigner:   false,
				IsWritable: false,
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
				PublicKey:  solana.SysVarRentPubkey,
				IsSigner:   false,
				IsWritable: false,
			},
		},
		data:      []byte{0},
		programID: solana.SPLAssociatedTokenAccountProgramID,
	}
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
