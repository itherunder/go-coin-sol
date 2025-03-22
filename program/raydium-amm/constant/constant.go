package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/itherunder/go-coin-sol/type"
)

var (
	Raydium_AMM_Program = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("HWy1jotHpo6UqeQxx49dpYYdQB8wj9Qk9MdxwjLvDHB8"),
	}
	Raydium_Authority_V4 = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("DbQqP6ehDYmeYjcBaMRuA8tAJY1EjDUz9DpwSLjaQqfC"),
	}

	Platform_Raydium_AMM type_.DexPlatform = "raydium_amm"
)

const (
	Raydium_Buy_Unit_Limit  = 90000
	Raydium_Sell_Unit_Limit = 90000
)
