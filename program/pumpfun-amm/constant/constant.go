package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/itherunder/go-coin-sol/type"
)

var (
	Pumpfun_AMM_Program = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"),
	}
	Event_Authority = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("GS4CU59F31iL7aR2Q8zVS8DRrcRnXX1yjQ66TqNVQnaR"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("GS4CU59F31iL7aR2Q8zVS8DRrcRnXX1yjQ66TqNVQnaR"),
	}

	Global_Config = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("ADyA8hdefvWN2dbGGWFotbzWxrAvLW83WG6QCVXvJKqw"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("ADyA8hdefvWN2dbGGWFotbzWxrAvLW83WG6QCVXvJKqw"),
	}

	Protocol_Fee_Recipient = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV"),
	}

	Protocol_Fee_Recipient_WSOL_Token_Account = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("94qWNrtmfn42h3ZjUZwWvK1MEo9uVmmrBPd2hpNjYDjb"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("94qWNrtmfn42h3ZjUZwWvK1MEo9uVmmrBPd2hpNjYDjb"),
	}

	Platform_Pumpfun_AMM type_.DexPlatform = "pumpfun_amm"
)

const (
	Pumpfun_Amm_Buy_Unit_Limit  = 100000
	Pumpfun_Amm_Sell_Unit_Limit = 90000
)
