package type_

import (
	"github.com/gagliardetto/solana-go"
)

type SwapV2Keys struct {
	// Token Program
	// Token 2022 Program
	// Memo Program v2
	// User address
	PairAddress solana.PublicKey
	MintA       solana.PublicKey
	MintB       solana.PublicKey
	// Token Owner Account A
	VaultA solana.PublicKey
	// Token Owner Account B
	VaultB     solana.PublicKey
	TickArray0 solana.PublicKey
	TickArray1 solana.PublicKey
	TickArray2 solana.PublicKey
	Oracle     solana.PublicKey

	Vaults map[solana.PublicKey]solana.PublicKey
}

type SwapKeys struct {
	// Token Program
	// User address
	PairAddress solana.PublicKey
	// Token Owner Account A
	VaultA solana.PublicKey
	// Token Owner Account B
	VaultB     solana.PublicKey
	TickArray0 solana.PublicKey
	TickArray1 solana.PublicKey
	TickArray2 solana.PublicKey
	Oracle     solana.PublicKey

	MintA  solana.PublicKey
	MintB  solana.PublicKey
	Vaults map[solana.PublicKey]solana.PublicKey
}
