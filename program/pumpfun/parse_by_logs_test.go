package pumpfun

import (
	"context"
	"fmt"
	"os"
	"testing"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/itherunder/go-coin-sol/constant"
	type_ "github.com/itherunder/go-coin-sol/program/pumpfun/type"
	go_test_ "github.com/pefish/go-test"
)

func TestParseCreateByLogs(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3eFnYcDuT9hTD38kF7DSQE9DurRPePsrj4D6vvLQE2Pq33kBfgEvHYDUYxjtAmVsSzm1xDkzWBGSAAFd6Bpx9y5n"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	d := ParseCreateByLogs(rpc.MainNetBeta, getTransactionResult.Meta.LogMessages)
	if d != nil {
		fmt.Printf(
			"<%s> <TokenAddress: %s> <UserAddress: %s> <URI: %s>\n",
			d.Symbol,
			d.TokenAddress,
			d.UserAddress,
			d.URI,
		)
	}
}

func TestParseSwapByLogs(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("22yrnjBaqiXSFJcaovcR5zfNQw2WtuDZ1kBfiTDj6uog67Fbc85bTxcD67F7QpNG1oefWCgZi7NeY57N4JHf19Wq"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	swaps := ParseSwapByLogs(rpc.MainNetBeta, getTransactionResult.Meta.LogMessages)
	for _, swap := range swaps {
		extraDatas := swap.ExtraDatas.(*type_.ExtraDatasType)
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<OutputAddress: %s>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<ReserveSOLAmountWithDecimals: %d>
<ReserveTokenAmountWithDecimals: %d>
<Timestamp: %d>
`,
			swap.UserAddress,
			swap.InputAddress,
			swap.OutputAddress,
			swap.InputAmountWithDecimals,
			swap.OutputAmountWithDecimals,
			extraDatas.ReserveSOLAmountWithDecimals,
			extraDatas.ReserveTokenAmountWithDecimals,
			extraDatas.Timestamp,
		)
	}
}

func TestIsRemoveLiqByLogs(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3SQiLy2biw1R6MZiPDyHVHo3QdHhAKNSZ5Z3RrKHxyV18vtdkm6xR88ygCqWENGKG8b1MM8tqTXwYt7SY21hY5VA"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	is := IsRemoveLiqByLogs(rpc.MainNetBeta, getTransactionResult.Meta.LogMessages)
	fmt.Println(is)
}

func TestIsAddLiqByLogs(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3wMndc1qNoxiQYDSThuWcNUnFdUw46kpNsVmS2siePbRpVoXWhKZHbuUBQT789jxdvfg4HAigJKqarSxN53wgE5F"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	is := IsAddLiqByLogs(rpc.MainNetBeta, getTransactionResult.Meta.LogMessages)
	fmt.Println(is)
}
