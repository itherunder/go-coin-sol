package constant

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	type_ "github.com/pefish/go-coin-sol/type"
)

var (
	Pumpfun_Program = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"),
	}
	Pumpfun_Event_Authority      = solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")
	Pumpfun_Raydium_Migration    = solana.MustPublicKeyFromBase58("39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg")
	Pumpfun_Token_Mint_Authority = solana.MustPublicKeyFromBase58("TSLvdd1pWpHVjahSpsvCXUbgwsL3JAcvokwaKt1eokM")

	Global = solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")

	Fee_Receiver = map[rpc.Cluster]solana.PublicKey{
		rpc.MainNetBeta: solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"),
		rpc.DevNet:      solana.MustPublicKeyFromBase58("68yFSZxzLWJXkxxRGydZ63C6mHx1NLEDWmwN9Lb5yySg"),
	}

	Platform_Pumpfun type_.DexPlatform = "pumpfun"
)

const (
	Pumpfun_Token_Decimals  = 6
	Pumpfun_Buy_Unit_Limit  = 80000
	Pumpfun_Sell_Unit_Limit = 60000
)
