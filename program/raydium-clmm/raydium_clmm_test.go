package raydium_clmm

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/constant"
	raydium_clmm_type "github.com/itherunder/go-coin-sol/program/raydium-clmm/type"
	type_ "github.com/itherunder/go-coin-sol/type"
	go_format "github.com/pefish/go-format"
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
	swapKeys := raydium_clmm_type.SwapV2Keys{
		PairAddress: solana.MustPublicKeyFromBase58("8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj"),
		Vaults: map[solana.PublicKey]solana.PublicKey{
			solana.SolMint: solana.MustPublicKeyFromBase58("6P4tvbzRY6Bh3MiWDHuLqyHywovsRwRpfskPvyeSoHsz"),
			tokenAddress:   solana.MustPublicKeyFromBase58("6mK4Pxs6GhwnessH7CvPivqDYauiHZmAdbEFDpXFk9zt"),
		},
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
		swapKeys.PairAddress,
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

func TestParseSwapTxByParsedTx(t *testing.T) {
	// return
	// 3jNs8EM3btxGUDP5AyvYX8NRgPcZpPqpRry6Pb6YLdR1CWAZTEcfo2PZzE3VWTqKYGkxqqyuNKdezxedX7j7QPYk
	// 4Joi4gD36KPcsHoqPjXhQWPKSrmTdJTNvQbtbJJbbbY8RwAjtLmpXFVS4s4WbRDtbD6fjs8LQMUKk6xbmvcfBoVp
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4Joi4gD36KPcsHoqPjXhQWPKSrmTdJTNvQbtbJJbbbY8RwAjtLmpXFVS4s4WbRDtbD6fjs8LQMUKk6xbmvcfBoVp"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	for _, swap := range r.Swaps {
		extraDatas := swap.ExtraDatas.(*raydium_clmm_type.ExtraDatasType)
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<OutputAddress: %s>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<InputVault: %s>
<OutputVault: %s>
<RemainAccounts: %s>	
<ReserveInputWithDecimals: %d>
<ReserveOutputWithDecimals: %d>
`,
			swap.UserAddress,
			swap.InputAddress,
			swap.OutputAddress,
			swap.InputAmountWithDecimals,
			swap.OutputAmountWithDecimals,
			swap.InputVault,
			swap.OutputVault,
			go_format.ToString(swap.ParsedKeys.(*raydium_clmm_type.SwapKeys).RemainAccounts),
			extraDatas.ReserveInputWithDecimals,
			extraDatas.ReserveOutputWithDecimals,
		)
	}
}
