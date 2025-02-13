package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/pefish/go-coin-sol/type"
)

var (
	Meteora_DLMM = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH"),
	}

	Meteora_DLMM_Event_Authority = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("D1ZN9Wj1fRSUQfCjhvnu1hqDMT7hzjzBBpi12nVniYD6"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH"),
	}

	Platform_Meteora_DLMM type_.DexPlatform = "meteora_dlmm"
)
