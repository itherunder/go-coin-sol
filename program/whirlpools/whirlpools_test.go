package whirlpools

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
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
	// xomHXr1pPcYJSuyEKoFu49zVYJAzGMHHzvoNDrA9Pb4zeSa26m3hpMoSjxTuMtT3vyMk7e3oAqLtzHjYBStPZ2B
	// 2HsKsrpUS11vrESJAbUgRGUbjMf4Tgxkxp4VV1D6bYFKyfJb8CfjvqUzpSg7i2538edciEzKH41Q7KvwJQ9owcLy
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("2HsKsrpUS11vrESJAbUgRGUbjMf4Tgxkxp4VV1D6bYFKyfJb8CfjvqUzpSg7i2538edciEzKH41Q7KvwJQ9owcLy"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	for _, swap := range r.Swaps {
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<OutputAddress: %s>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<InputVault: %s>
<OutputVault: %s>
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
			swap.ReserveInputWithDecimals,
			swap.ReserveOutputWithDecimals,
		)
	}
}
