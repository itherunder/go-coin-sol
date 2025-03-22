package util

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/itherunder/go-coin-sol/constant"
	i_logger "github.com/itherunder/go-interface/i-logger"
	go_test_ "github.com/itherunder/go-test"
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

func TestGetFeeInfoFromTx(t *testing.T) {
	// return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("2BWpWYCiHiRSy4gTHDRc32HbrbrZ8WtTqMhM21TQSzPhtRbCcqmTdUCfBeSqUDFRLaquDQ3o1Nd5hFgEv4o5RF5m"),
		&rpc.GetTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	tx, err := getTransactionResult.Transaction.GetTransaction()
	go_test_.Equal(t, nil, err)
	r, err := GetFeeInfoFromTx(getTransactionResult.Meta, tx)
	go_test_.Equal(t, nil, err)
	fmt.Printf("%d, %d, %d\n", r.BaseFeeWithDecimals, r.PriorityFeeWithDecimals, r.TotalFeeWithDecimals)
}

func TestGetComputeUnitPriceFromHelius(t *testing.T) {
	// return
	r, err := GetComputeUnitPriceFromHelius(
		&i_logger.DefaultLogger,
		os.Getenv("KEY"),
		[]string{
			"CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM",
			"EJV5M9pjYxcHwJBNZqCkig47y5H795cinwhHJwMtyRvZ",
			"4h8kyVuaSfXrrgqUzejeuLKYVUAnNktwMj1PFASt7NrD",
			"HqpgLHJ5nyACChz4SQJWvwi9oVBxMkUJQXa4woNu1DL9",
			"A3nQhBN4NTjEvtbKaSLZu6kGFrgYWkWPyR9VLymbqZ7R",
		},
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(r)
}

func TestGetReserves(t *testing.T) {
	// return
	r1, r2, err := GetReserves(
		client,
		solana.MustPublicKeyFromBase58("8GQnRr8BpAq2ad1vsWtjVz3xwirg26pUQhWFQnm17eXT"),
		solana.MustPublicKeyFromBase58("8xWMKfLguLv6TTZsQqvRq23G7rMwwxV5bE1BHJ9rrS4r"),
	)
	go_test_.Equal(t, nil, err)
	fmt.Printf(
		`
<r1.AmountWithDecimals: %d>
<r1.Decimals: %d>
<r2.AmountWithDecimals: %d>
<r2.Decimals: %d>
`,
		r1.AmountWithDecimals,
		r1.Decimals,
		r2.AmountWithDecimals,
		r2.Decimals,
	)
}
