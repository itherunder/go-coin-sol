package util

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

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

func TestGetDiscriminator(t *testing.T) {
	r := GetDiscriminator("global", "swap_v2")
	fmt.Println(r)
}
