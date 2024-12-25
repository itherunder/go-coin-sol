package raydium

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	go_test_ "github.com/pefish/go-test"
)

func TestParseSwapTx(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("5eUYrCrWVYpnAE1NHsDw9aPgzbtMofe8GXPZc3rDD8Xm3qnQqoMP9piYwdb8q3NeXDeyXnSgqWuaPBvsQ4VTooTP"),
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
	for _, swap := range r.Swaps {
		fmt.Printf(
			"<UserAddress: %s> <%s> <TokenAddress: %s> <%s sol> <TokenAmount: %s> <UserBalance: %s> <UserTokenBalance: %s>\n",
			swap.UserAddress,
			swap.Type,
			swap.TokenAddress,
			swap.SOLAmount,
			swap.TokenAmount,
			swap.UserBalance,
			swap.UserTokenBalance,
		)
	}

}
