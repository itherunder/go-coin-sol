package constant

import solana "github.com/gagliardetto/solana-go"

var (
	Pumpfun_Program           = solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
	Pumpfun_Event_Authority   = solana.MustPublicKeyFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1")
	Pumpfun_Raydium_Migration = solana.MustPublicKeyFromBase58("39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg")

	Global       = solana.MustPublicKeyFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf")
	Fee_Receiver = solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM")
)

const (
	Pumpfun_Token_Decimals  = 6
	Pumpfun_Buy_Unit_Limit  = 65000
	Pumpfun_Sell_Unit_Limit = 40000
)
