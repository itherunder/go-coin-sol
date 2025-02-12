package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	WhirlpoolsProgram = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH"),
	}
)
