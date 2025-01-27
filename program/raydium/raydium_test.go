package raydium

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
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
	r, err := ParseSwapTx(rpc.MainNetBeta, getTransactionResult.Meta, tx, true)
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
		solana.MustSignatureFromBase58("4isBHXQ6y9CPvuJetwHcDZWypB952prBHn9ZGokx5pTQUrPKd5241uxyL5SQNQRUEFRefXLRWuZVzaNAaq2JqPAt"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction, false)
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

func TestGetSwapInstructions(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("82zGj6ee2ocCMBH1mogyNLr8paUoai45GYVa42QYfzPz")
	tokenDecimals := 9
	coinIsSOL := false
	raydiumSwapKeys := raydium_type_.RaydiumSwapKeys{
		AmmAddress:                  solana.MustPublicKeyFromBase58("HfzUC934vUPc7E8G7YtbgtaWrrjKhraDF4ZEZ8A6gsYA"),
		PoolCoinTokenAccountAddress: solana.MustPublicKeyFromBase58("7ZuXkdD9dTYXJr38W2KGdDLjssN61VxkWzANkFLfeQKe"),
		PoolPcTokenAccountAddress:   solana.MustPublicKeyFromBase58("8ErAcSyRyWg5xDhzR28fpoA8EPDDUqQaqmcz2pSAZX3J"),
	}
	solAmount, tokenAmount, err := GetReserves(
		client,
		raydiumSwapKeys.PoolCoinTokenAccountAddress,
		raydiumSwapKeys.PoolPcTokenAccountAddress,
		coinIsSOL,
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		uint64(4*math.Pow(10, float64(tokenDecimals))),
		raydiumSwapKeys,
		true,
		solAmount.AmountWithDecimals,
		tokenAmount.AmountWithDecimals,
		50,
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(swapInstructions)
}
