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
	// return
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4TRPLs5TzMqo4VmayEFiHGWQyoTyAgvbViqqiJoQWGTR6D7oDjPw4NRWtVUbeJCjpg6T6v5ND4NP8Ms8EPndiQuu"),
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
			"<UserAddress: %s> <%s> <TokenAddress: %s> <%d sol> <TokenAmount: %d> <UserBalance: %d -> %d> <UserTokenBalance: %d -> %d>\n",
			swap.UserAddress,
			swap.Type,
			swap.TokenAddress,
			swap.SOLAmountWithDecimals,
			swap.TokenAmountWithDecimals,
			swap.BeforeUserBalanceWithDecimals,
			swap.UserBalanceWithDecimals,
			swap.BeforeUserTokenBalanceWithDecimals,
			swap.UserTokenBalanceWithDecimals,
		)
	}

}

func TestParseSwapTxByParsedTx(t *testing.T) {
	// return
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4TRPLs5TzMqo4VmayEFiHGWQyoTyAgvbViqqiJoQWGTR6D7oDjPw4NRWtVUbeJCjpg6T6v5ND4NP8Ms8EPndiQuu"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	for _, swap := range r.Swaps {
		fmt.Printf(
			"<UserAddress: %s> <%s> <TokenAddress: %s> <%d sol> <TokenAmount: %d> <UserBalance: %d -> %d> <UserTokenBalance: %d -> %d>\n",
			swap.UserAddress,
			swap.Type,
			swap.TokenAddress,
			swap.SOLAmountWithDecimals,
			swap.TokenAmountWithDecimals,
			swap.BeforeUserBalanceWithDecimals,
			swap.UserBalanceWithDecimals,
			swap.BeforeUserTokenBalanceWithDecimals,
			swap.UserTokenBalanceWithDecimals,
		)
	}

}
