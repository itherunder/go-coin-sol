package pumpfun

import (
	"context"
	"fmt"
	"os"
	"testing"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
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
		solana.MustSignatureFromBase58("F5ahQP2qDcktN7MKW3mQJJY7dM279naRNPVZE9z9fnTyUPjsNX7kKok7vUYZe5LjuyYVgkXwYpZXMxZxfiVyAaf"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	d, err := ParseCreateByLogs(getTransactionResult.Meta.LogMessages)
	go_test_.Equal(t, nil, err)
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
		solana.MustSignatureFromBase58("3FhAfwZts7di6LwtTY86rVGprB1hvsMtNrpfmt95UxfvH4LSZsn2fjMxuekmm7sx6ZKvxuwzWQhYc7yZdrb2r2f9"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	swaps, err := ParseSwapByLogs(getTransactionResult.Meta.LogMessages)
	go_test_.Equal(t, nil, err)
	for _, swapData := range swaps {
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
	is, err := IsRemoveLiqByLogs(getTransactionResult.Meta.LogMessages)
	go_test_.Equal(t, nil, err)
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
		solana.MustSignatureFromBase58("4HYw39TArnMWivKBxjNBA4jMJEJ6ckzUDoVC5Au9zk1qUzYRa7yygxPscvveZAKocftbw7CfmX36hfg7qw2PUguD"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	is, err := IsAddLiqByLogs(getTransactionResult.Meta.LogMessages)
	go_test_.Equal(t, nil, err)
	fmt.Println(is)
}
