package raydium

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

func TestParseSwapTx(t *testing.T) {
	// return
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	client := rpc.New(url)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("28qwqwrnbKxKn9Ek5Wi4t5HaYpH3G9XVYS5jE28FqgomU2eZPpRBSZkXZ6KapdyRBXTB26Z31deUkpbc69YMnSQm"),
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
			"<UserAddress: %s> <%s> <TokenAddress: %s> <%d sol> <TokenAmount: %d> <UserBalance: %d -> %d> <UserTokenBalance: %d>\n",
			swap.UserAddress,
			swap.Type,
			swap.TokenAddress,
			swap.SOLAmountWithDecimals,
			swap.TokenAmountWithDecimals,
			swap.BeforeUserBalanceWithDecimals,
			swap.UserBalanceWithDecimals,
			swap.UserTokenBalanceWithDecimals,
		)
	}

}
