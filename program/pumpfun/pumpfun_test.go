package pumpfun

import (
	"context"
	"fmt"
	"testing"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

func TestParseSwapTx(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
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
			"[Swap] <%s> <SOLAmount: %s> <TokenAmount: %s> <ReserveSOLAmount: %s> <ReserveTokenAmount: %s>\n",
			swapData.Type,
			swapData.SOLAmount,
			swapData.TokenAmount,
			swapData.ReserveSOLAmount,
			swapData.ReserveTokenAmount,
		)
	}
}

func TestParseSwapByLogs(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
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
			"[Swap] <%s> <SOLAmount: %s> <TokenAmount: %s> <ReserveSOLAmount: %s> <ReserveTokenAmount: %s>\n",
			swapData.Type,
			swapData.SOLAmount,
			swapData.TokenAmount,
			swapData.ReserveSOLAmount,
			swapData.ReserveTokenAmount,
		)
	}
}

func TestParseCreateByLogs(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
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
	fmt.Printf(
		"<%s> <TokenAddress: %s> <UserAddress: %s> <URI: %s>\n",
		d.Symbol,
		d.TokenAddress,
		d.UserAddress,
		d.URI,
	)
}

func TestParseTx(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("ystcu4rAXbEeaEFDAPc3CSpufZu2t92SzM8x8xBmQWbjYymxpBqyTsNJQuJ4twWpuvoisYzqqRBBp5Rhea1Vkb7"),
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
				"[Swap] <SOLAmount: %s> <TokenAmount: %s> <UserBalance: %s>\n",
				swapData.SOLAmount,
				swapData.TokenAmount,
				r.SwapTxData.UserBalance,
			)
		}
	}
}

func TestURIInfo(t *testing.T) {
	r, err := URIInfo(&i_logger.DefaultLogger, "https://ipfs.io/ipfs/QmVSKrX4XxUgHMCp2wnmE3VmnK3fCyZsSqFiVGQZowiM1c")
	go_test_.Equal(t, nil, err)
	fmt.Println("aa", r.Twitter)
}
