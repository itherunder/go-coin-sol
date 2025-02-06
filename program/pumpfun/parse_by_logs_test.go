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
		solana.MustSignatureFromBase58("3eFnYcDuT9hTD38kF7DSQE9DurRPePsrj4D6vvLQE2Pq33kBfgEvHYDUYxjtAmVsSzm1xDkzWBGSAAFd6Bpx9y5n"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	d := ParseCreateByLogs(getTransactionResult.Meta.LogMessages)
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
	swaps := ParseSwapByLogs(getTransactionResult.Meta.LogMessages)
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
	is := IsRemoveLiqByLogs(getTransactionResult.Meta.LogMessages)
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
		solana.MustSignatureFromBase58("2yP7DNDVMRP9VeGLQKKp6RNnc8XJxaRisv1ViXuawK4XkeVjvpW3gkN5XdMimhMJmE12PCEGBvDLHqKBq1CeFybJ"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	is := IsAddLiqByLogs(rpc.MainNetBeta, getTransactionResult.Meta.LogMessages)
	fmt.Println(is)
}
