package raydium_amm

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/constant"
	raydium_amm_type "github.com/itherunder/go-coin-sol/program/raydium-amm/type"
	raydium_type_ "github.com/itherunder/go-coin-sol/program/raydium-amm/type"
	type_ "github.com/itherunder/go-coin-sol/type"
	"github.com/itherunder/go-coin-sol/util"
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

func TestParseSwapTxByParsedTx(t *testing.T) {
	// return
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3TEAsdxPBCwmLnT4ZfzgSHa3YqfzD1aRXb8vqCPG6AxtBzQVzjKT8uk4FHVUkry3GkLitEoZqjpQrnWNR4M5ptXq"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)

	for _, swap := range r.Swaps {
		extraDatas := swap.ExtraDatas.(*raydium_amm_type.ExtraDatasType)
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<InputDecimals: %d>
<OutputAddress: %s>
<OutputDecimals: %d>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<InputVault: %s>
<OutputVault: %s>
<ReserveInputWithDecimals: %d>
<ReserveOutputWithDecimals: %d>
`,
			swap.UserAddress,
			swap.InputAddress,
			swap.InputDecimals,
			swap.OutputAddress,
			swap.OutputDecimals,
			swap.InputAmountWithDecimals,
			swap.OutputAmountWithDecimals,
			swap.InputVault,
			swap.OutputVault,
			extraDatas.ReserveInputWithDecimals,
			extraDatas.ReserveOutputWithDecimals,
		)
	}

}

func TestGetSwapInstructions(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("82zGj6ee2ocCMBH1mogyNLr8paUoai45GYVa42QYfzPz")
	tokenDecimals := 9
	raydiumSwapKeys := raydium_type_.SwapKeys{
		AmmAddress: solana.MustPublicKeyFromBase58("HfzUC934vUPc7E8G7YtbgtaWrrjKhraDF4ZEZ8A6gsYA"),
		CoinMint:   tokenAddress,
		PCMint:     solana.SolMint,
		Vaults: map[solana.PublicKey]solana.PublicKey{
			solana.SolMint: solana.MustPublicKeyFromBase58("8ErAcSyRyWg5xDhzR28fpoA8EPDDUqQaqmcz2pSAZX3J"),
			tokenAddress:   solana.MustPublicKeyFromBase58("7ZuXkdD9dTYXJr38W2KGdDLjssN61VxkWzANkFLfeQKe"),
		},
	}
	solAmount, tokenAmount, err := util.GetReserves(
		client,
		raydiumSwapKeys.Vaults[solana.SolMint],
		raydiumSwapKeys.Vaults[tokenAddress],
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		uint64(4*math.Pow(10, float64(tokenDecimals))),
		raydiumSwapKeys,
		true,
		solAmount.AmountWithDecimals,
		tokenAmount.AmountWithDecimals,
		50,
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(swapInstructions)
}

func TestParseAddLiqTxByParsedTx(t *testing.T) {
	// return
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("6eEwKoj8UdEDg2hcnhCACiaPsBfmu34fo7c2fs5F5xEmpUYakSQfiSR2kzbNNXfAhhXq5aBGmicsj5UscG3djhx"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseAddLiqTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, r == nil)
	fmt.Printf(
		`
<TokenAddress: %s>
<AMMAddress: %s>
<PoolCoinTokenAccount: %s>
<PoolPcTokenAccount: %s>
<CoinMint: %s>
<PCMint: %s>
		`,
		r.TokenAddress,
		r.AmmAddress.String(),
		r.Vaults[r.CoinMint].String(),
		r.Vaults[r.PCMint].String(),
		r.CoinMint.String(),
		r.PCMint.String(),
	)
}
