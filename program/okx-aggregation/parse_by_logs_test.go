package pumpfun

import (
	"context"
	"fmt"
	"os"
	"testing"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	go_test_ "github.com/pefish/go-test"
)

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
		solana.MustSignatureFromBase58("4g85rwNon9u5NVdL3JQD19n4Akuo8VV3ZUNcB35u7cTzxjEUpFWbmS4W2juHwfaTEcT78SvPvrvWDvguAE1MaetZ"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	swapData := ParseSwapByLogs(getTransactionResult.Meta.LogMessages)
	fmt.Printf(
		"[Swap] <%s> <SOLAmount: %d> <TokenAmount: %d> <UserTokenBalanceWithDecimals: %d>\n",
		swapData.Type,
		swapData.SOLAmountWithDecimals,
		swapData.TokenAmountWithDecimals,
		swapData.UserTokenBalanceWithDecimals,
	)
}
