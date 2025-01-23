package pumpfun

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	i_logger "github.com/pefish/go-interface/i-logger"
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

func TestParseSwapTx(t *testing.T) {
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3FhAfwZts7di6LwtTY86rVGprB1hvsMtNrpfmt95UxfvH4LSZsn2fjMxuekmm7sx6ZKvxuwzWQhYc7yZdrb2r2f9"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	tx, err := getTransactionResult.Transaction.GetTransaction()
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTx(getTransactionResult.Meta, tx)
	go_test_.Equal(t, nil, err)
	for _, swapData := range r.Swaps {
		fmt.Printf(
			"[Swap] <%s> <SOLAmount: %d> <TokenAmount: %d> <ReserveSOLAmount: %d> <ReserveTokenAmount: %d>\n",
			swapData.Type,
			swapData.SOLAmountWithDecimals,
			swapData.TokenAmountWithDecimals,
			swapData.ReserveSOLAmountWithDecimals,
			swapData.ReserveTokenAmountWithDecimals,
		)
	}
}

func TestParseTx(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4jj1WgBN8QYP7pDiazyVXwiwnJQnBVKJM7NpXHEMJiqnu6HfYitBgtEd9hnxtYkpvMjTDsUbgqtFWxnw63J42UdP"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	tx, err := getTransactionResult.Transaction.GetTransaction()
	go_test_.Equal(t, nil, err)
	r, err := ParseTx(getTransactionResult.Meta, tx)
	go_test_.Equal(t, nil, err)
	if r.CreateTxData != nil {
		fmt.Printf("[Create] <%s> <%s>\n", r.CreateTxData.Name, r.CreateTxData.Symbol)
	}

	if r.RemoveLiqTxData != nil {
		fmt.Printf("[RemoveLiq] %+v\n", r.RemoveLiqTxData)
	}

	if r.AddLiqTxData != nil {
		fmt.Printf("[AddLiq] %+v\n", r.AddLiqTxData)
	}

	if r.SwapTxData != nil {
		for _, swapData := range r.SwapTxData.Swaps {
			fmt.Printf(
				"[Swap] <%s> <SOLAmount: %d> <TokenAmount: %d> <UserBalance: %d>\n",
				swapData.Type,
				swapData.SOLAmountWithDecimals,
				swapData.TokenAmountWithDecimals,
				r.SwapTxData.UserBalanceWithDecimals,
			)
		}
	}
}

func TestURIInfo(t *testing.T) {
	r, err := URIInfo(&i_logger.DefaultLogger, "https://ipfs.io/ipfs/QmVSKrX4XxUgHMCp2wnmE3VmnK3fCyZsSqFiVGQZowiM1c")
	go_test_.Equal(t, nil, err)
	fmt.Println("aa", r.Twitter)
}

func TestGetBondingCurveData(t *testing.T) {
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	tokenAddressObj := solana.MustPublicKeyFromBase58("7PAaQ8UxYYPksnyxcKFP44Pm4FwFCix4ammGf5P3bK79")
	r, err := GetBondingCurveData(client, &tokenAddressObj, nil)
	go_test_.Equal(t, nil, err)
	fmt.Println(r)
}

func TestGenerateTokenURI(t *testing.T) {
	// return
	r, err := GenerateTokenURI(&GenerateTokenURIDataType{
		Name:        "testcoin",
		Symbol:      "TEST",
		Description: "test test.",
		File:        nil,
		Twitter:     "https://x.com",
		Website:     "https://x.com",
		Telegram:    "https://tg.com",
	})
	go_test_.Equal(t, nil, err)
	fmt.Printf("%#v\n", r)
}

func TestGenePumpfunWallet(t *testing.T) {
	// return
	r, err := GenePumpfunWallet(2 * time.Minute)
	go_test_.Equal(t, nil, err)
	fmt.Println(r.PublicKey().String())
}

func TestParseRemoveLiqTxByParsedTx(t *testing.T) {
	// return
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("Dw9KGZqJ9CAB5PWe89r9VoDTfW97Hw49EN6GjLZ5W4RLrzx8i2sLny4uBqXHqQkZNZt8FCGcoTonqrj6JrkZkDm"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseRemoveLiqTxByParsedTx(getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, r == nil)
	fmt.Printf(
		"[RemoveLiq] <%s> <BondingCurveAddress: %s>\n",
		r.TokenAddress,
		r.BondingCurveAddress.String(),
	)
}

func TestParseAddLiqTxByParsedTx(t *testing.T) {
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("44sEeJxeoZiZDoT4dakF6kKuynenFgWYevzwuzMqGsarvxd5bQKYcMzZWxh1kEnZxd8uiAKAjs8YfAXCoM2pAGm4"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseAddLiqTxByParsedTx(getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, r == nil)
	fmt.Printf(
		"[AddLiq] <%s> <AMMAddress: %s> <PoolCoinTokenAccount: %s> <PoolPcTokenAccount: %s>\n",
		r.TokenAddress,
		r.AMMAddress.String(),
		r.PoolCoinTokenAccount.String(),
		r.PoolPcTokenAccount.String(),
	)
}
