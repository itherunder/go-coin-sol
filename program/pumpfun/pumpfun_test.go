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
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("2pufWS8EvN8qQttHcChYCp8AoCaq5Fo1Drj1zNH9ejGDQhgJerssw6qF2krGMbDR6uR2ptM3geWBswwpstmW8yXs"),
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
	fmt.Println(r)
}

func TestParseTx(t *testing.T) {
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4zGZfarZKsp6SNkWv2YXzeTBCuFKUAaHbGeoTTU64bkXBCU6Km4ZWrscW9frTRofQybneH33siPnhwtHyuxF5AJ7"),
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
				"[Swap] <SOLAmount: %s> <TokenAmount: %s> <UserBalance: %s> <UserTokenBalance: %s>\n",
				swapData.SOLAmount,
				swapData.TokenAmount,
				swapData.UserBalance,
				swapData.UserTokenBalance,
			)
		}
	}
}

func TestURIInfo(t *testing.T) {
	r, err := URIInfo(&i_logger.DefaultLogger, "https://ipfs.io/ipfs/QmVSKrX4XxUgHMCp2wnmE3VmnK3fCyZsSqFiVGQZowiM1c")
	go_test_.Equal(t, nil, err)
	fmt.Println("aa", r.Twitter)
}
