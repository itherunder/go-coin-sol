package raydium_amm

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium-amm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
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
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4isBHXQ6y9CPvuJetwHcDZWypB952prBHn9ZGokx5pTQUrPKd5241uxyL5SQNQRUEFRefXLRWuZVzaNAaq2JqPAt"),
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

func TestGetSwapInstructions(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("82zGj6ee2ocCMBH1mogyNLr8paUoai45GYVa42QYfzPz")
	tokenDecimals := 9
	raydiumSwapKeys := raydium_type_.SwapKeys{
		AmmAddress:                  solana.MustPublicKeyFromBase58("HfzUC934vUPc7E8G7YtbgtaWrrjKhraDF4ZEZ8A6gsYA"),
		PoolCoinTokenAccountAddress: solana.MustPublicKeyFromBase58("7ZuXkdD9dTYXJr38W2KGdDLjssN61VxkWzANkFLfeQKe"),
		PoolPcTokenAccountAddress:   solana.MustPublicKeyFromBase58("8ErAcSyRyWg5xDhzR28fpoA8EPDDUqQaqmcz2pSAZX3J"),
	}
	solAmount, tokenAmount, err := util.GetReserves(
		client,
		raydiumSwapKeys.PoolPcTokenAccountAddress,
		raydiumSwapKeys.PoolCoinTokenAccountAddress,
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

func TestParseAddLiqTxByParsedTx(t *testing.T) {
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("44sEeJxeoZiZDoT4dakF6kKuynenFgWYevzwuzMqGsarvxd5bQKYcMzZWxh1kEnZxd8uiAKAjs8YfAXCoM2pAGm4"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseAddLiqTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, r == nil)
	fmt.Printf(
		"[AddLiq] <%s> <AMMAddress: %s> <PoolCoinTokenAccount: %s> <PoolPcTokenAccount: %s>\n",
		r.TokenAddress,
		r.AMMAddress.String(),
		r.PoolCoinTokenAccount.String(),
		r.PoolPcTokenAccount.String(),
	)
}
