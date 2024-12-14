package instruction

import (
	solana "github.com/gagliardetto/solana-go"
)

type CreateIdempotentInstruction struct {
	accounts  []*solana.AccountMeta
	data      []byte
	programID solana.PublicKey
}

func NewCreateIdempotentInstruction(
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
		data:      []byte{1},
		programID: solana.SPLAssociatedTokenAccountProgramID,
	}
}

func (t *CreateIdempotentInstruction) Accounts() []*solana.AccountMeta {
	return t.accounts
}

func (t *CreateIdempotentInstruction) ProgramID() solana.PublicKey {
	return t.programID
}

func (t *CreateIdempotentInstruction) Data() ([]byte, error) {
	return t.data, nil
}
