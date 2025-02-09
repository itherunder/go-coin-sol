package raydium_clmm

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	raydium_clmm_type "github.com/pefish/go-coin-sol/program/raydium-clmm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	go_test_ "github.com/pefish/go-test"
)

var client *rpc.Client

func init() {
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client = rpc.New(url)
}

func TestGetSwapInstructions(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	tokenDecimals := 6
	swapKeys := raydium_clmm_type.SwapKeys{
		PoolIdAddress:    solana.MustPublicKeyFromBase58("8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj"),
		WSOLVault:        solana.MustPublicKeyFromBase58("6P4tvbzRY6Bh3MiWDHuLqyHywovsRwRpfskPvyeSoHsz"),
		TokenVault:       solana.MustPublicKeyFromBase58("6mK4Pxs6GhwnessH7CvPivqDYauiHZmAdbEFDpXFk9zt"),
		ObservationState: solana.MustPublicKeyFromBase58("3MsJXVvievxAbsMsaT6TS4i6oMitD9jazucuq3X234tC"),
		ExBitmapAccount:  solana.MustPublicKeyFromBase58("DoPuiZfJu7sypqwR4eiU7C5TMcmmiFoU4HaF5SoD8mRy"),
		RemainAccounts: []solana.PublicKey{
			solana.MustPublicKeyFromBase58("EWh7X48uss9YeikJjm2fEZnHeJAPVBst8QonbiHtHQxH"),
			solana.MustPublicKeyFromBase58("GpTybAYJ8899axRzSBLBEbHNhC1gnz1z7CtMQpC11x8N"),
			solana.MustPublicKeyFromBase58("9U7qaFspMpESCptyHZyPok5Bf2Hh97z8iMvcUbTEFuTH"),
		},
	}
	poolInfo, err := GetPoolInfo(
		client,
		swapKeys.PoolIdAddress,
	)
	go_test_.Equal(t, nil, err)
	fmt.Println("solAmount", poolInfo.SwapInAmountToken0.String())
	fmt.Println("tokenAmount", poolInfo.SwapInAmountToken1.String())

	swapInstructions, err := GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		uint64(1*math.Pow(10, float64(tokenDecimals))),
		swapKeys,
		true,
		poolInfo.SwapInAmountToken0.BigInt().Uint64(),
		poolInfo.SwapInAmountToken1.BigInt().Uint64(),
		50,
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(swapInstructions)
}

func TestGetPoolInfo(t *testing.T) {
	poolInfo, err := GetPoolInfo(
		client,
		solana.MustPublicKeyFromBase58("8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj"),
	)
	go_test_.Equal(t, nil, err)
	fmt.Println("token0", poolInfo.TokenMint0.String())
	fmt.Println("token1", poolInfo.TokenMint1.String())
	fmt.Println("solAmount", poolInfo.SwapInAmountToken0.String())
	fmt.Println("tokenAmount", poolInfo.SwapInAmountToken1.String())
}

func TestGetReserves(t *testing.T) {
	solReserve, tokenReserve, err := GetReserves(
		client,
		solana.MustPublicKeyFromBase58("8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj"),
	)
	go_test_.Equal(t, nil, err)
	fmt.Println("solReserve", solReserve.AmountWithDecimals)
	fmt.Println("tokenReserve", tokenReserve.AmountWithDecimals)
}
