package type_

import (
	"github.com/gagliardetto/solana-go"
)

type SwapKeys struct {
	PairAddress solana.PublicKey
	MintA       solana.PublicKey
	MintB       solana.PublicKey
	Vaults      map[solana.PublicKey]solana.PublicKey
}
