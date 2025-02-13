package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/pefish/go-coin-sol/type"
)

var (
	Raydium_Concentrated_Liquidity = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH"),
	}

	Platform_Raydium_CLMM type_.DexPlatform = "raydium_clmm"
)
