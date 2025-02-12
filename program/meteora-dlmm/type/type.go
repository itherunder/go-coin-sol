package type_

import (
	"github.com/gagliardetto/solana-go"
)

type SwapKeys struct {
	PairAddress    solana.PublicKey
	VaultX         solana.PublicKey
	VaultY         solana.PublicKey
	Oracle         solana.PublicKey
	RemainAccounts []solana.PublicKey
	XMint          solana.PublicKey
	YMint          solana.PublicKey
	Vaults         map[solana.PublicKey]solana.PublicKey
}
