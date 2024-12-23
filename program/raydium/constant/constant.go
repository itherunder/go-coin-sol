package constant

import solana "github.com/gagliardetto/solana-go"

var (
	Raydium_Liquidity_Pool_V4 = solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")
	Raydium_Authority_V4      = solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1")
)

const (
	Raydium_Buy_Unit_Limit  = 65000
	Raydium_Sell_Unit_Limit = 40000
)
