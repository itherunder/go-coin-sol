package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/itherunder/go-coin-sol/type"
)

var (
	SolFiProgram = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("SoLFiHG9TfgtdUXUjWAxi3LtvYuFyDLVhBWxdMZxyCe"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("HWy1jotHpo6UqeQxx49dpYYdQB8wj9Qk9MdxwjLvDHB8"),
	}

	Platform_Sol_Fi type_.DexPlatform = "sol_fi"
)
