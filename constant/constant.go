package constant

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	meteora_dlmm_constant "github.com/pefish/go-coin-sol/program/meteora-dlmm/constant"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	raydium_amm_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	raydium_clmm_constant "github.com/pefish/go-coin-sol/program/raydium-clmm/constant"
	sol_fi_constant "github.com/pefish/go-coin-sol/program/sol-fi/constant"
	whirlpools_constant "github.com/pefish/go-coin-sol/program/whirlpools/constant"
	type_ "github.com/pefish/go-coin-sol/type"
)

const (
	SOL_Decimals = 9
)

var (
	version = uint64(0)

	MaxSupportedTransactionVersion_0 = &version

	Dex_Platforms = map[type_.DexPlatform]map[rpc.Cluster]solana.PublicKey{
		pumpfun_constant.Platform_Pumpfun:           pumpfun_constant.Pumpfun_Program,
		raydium_amm_constant.Platform_Raydium_AMM:   raydium_amm_constant.Raydium_AMM_Program,
		raydium_clmm_constant.Platform_Raydium_CLMM: raydium_clmm_constant.Raydium_CLMM_Program,
		sol_fi_constant.Platform_Sol_Fi:             sol_fi_constant.SolFiProgram,
		whirlpools_constant.Platform_Whirlpools:     whirlpools_constant.WhirlpoolsProgram,
		meteora_dlmm_constant.Platform_Meteora_DLMM: meteora_dlmm_constant.Meteora_DLMM_Program,
	}
)
