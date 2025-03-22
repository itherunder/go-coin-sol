package pumpfun_amm

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/constant"
	pumpfun_amm_type "github.com/itherunder/go-coin-sol/program/pumpfun-amm/type"
	pumpfun_constant "github.com/itherunder/go-coin-sol/program/pumpfun/constant"
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
	// rWjZMva9ywMhNL2meJY1tU5UbV7i7aRutZGqeroYqdzkzp16poS5dzLLQ1KsetMQWmttZKQbKUoAyTtKnUcpPwF
	// 3ieyAgz19McF9sUqeEhXAvQQAvrrpaCk1Y7zmPgWFtG7Yosrb9QA9F6a4Dywhj34hVfpUzaEDMuAwX5PCnXevGSo
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3ieyAgz19McF9sUqeEhXAvQQAvrrpaCk1Y7zmPgWFtG7Yosrb9QA9F6a4Dywhj34hVfpUzaEDMuAwX5PCnXevGSo"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)

	for _, swap := range r.Swaps {
		extraDatas := swap.ExtraDatas.(*pumpfun_amm_type.ExtraDatasType)
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
	tokenDecimals := pumpfun_constant.Pumpfun_Token_Decimals
	tokenAmountWithDecimals := uint64(1000 * math.Pow(10, float64(tokenDecimals)))
	tokenAddress := solana.MustPublicKeyFromBase58("DP4MXhEhe9USfRr1pdDazEdqVftSVH95X7fAXG2epump")
	swapKeys := pumpfun_amm_type.SwapKeys{
		AmmAddress:        solana.MustPublicKeyFromBase58("4iucvyLyWumRqkL1WQXvcu1RyzPboczkKFjmEeR9WAN1"),
		BaseTokenAddress:  tokenAddress,
		QuoteTokenAddress: solana.SolMint,
	}
	baseVault, _ := swapKeys.BaseVault()
	quoteVault, _ := swapKeys.QuoteVault()
	solAmount, tokenAmount, err := util.GetReserves(
		client,
		quoteVault,
		baseVault,
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Buy,
		tokenAmountWithDecimals,
		swapKeys,
		false,
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
		solana.MustSignatureFromBase58("3wMndc1qNoxiQYDSThuWcNUnFdUw46kpNsVmS2siePbRpVoXWhKZHbuUBQT789jxdvfg4HAigJKqarSxN53wgE5F"),
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
<TxId: %s>
<AMMAddress: %s>
<BaseTokenAddress: %s>
<QuoteTokenAddress: %s>
<BaseTokenDecimals: %d>
<QuoteTokenDecimals: %d>
		`,
		r.TxId,
		r.AmmAddress.String(),
		r.BaseTokenAddress,
		r.QuoteTokenAddress,
		r.BaseTokenDecimals,
		r.QuoteTokenDecimals,
	)
}
