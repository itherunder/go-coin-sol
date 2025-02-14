package type_

import (
	"github.com/gagliardetto/solana-go"
)

type SwapV2Keys struct {
	PairAddress solana.PublicKey
	MintA       solana.PublicKey
	MintB       solana.PublicKey
	TickArray0  solana.PublicKey
	TickArray1  solana.PublicKey
	TickArray2  solana.PublicKey
	Oracle      solana.PublicKey

	Vaults map[solana.PublicKey]solana.PublicKey
}

type SwapKeys struct {
	PairAddress solana.PublicKey
	TickArray0  solana.PublicKey
	TickArray1  solana.PublicKey
	TickArray2  solana.PublicKey
	Oracle      solana.PublicKey

	MintA  solana.PublicKey
	MintB  solana.PublicKey
	Vaults map[solana.PublicKey]solana.PublicKey
}
