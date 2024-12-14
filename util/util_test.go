package go_coin_sol

import (
	"context"
	"fmt"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	go_test_ "github.com/pefish/go-test"
)

func TestGetFeeInfoFromTx(t *testing.T) {
	return
	endpoint := rpc.MainNetBeta_RPC
	client := rpc.New(endpoint)
	getTransactionResult, err := client.GetTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("6bojQmrqZdsqKLVW3ZZXK9wWnVr4G3N5jgJGaRsYtqHk6xsbpmzcsmXsMcpFGwdCEvSKon7SDAgDTSoVWa2QZ5N"),
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
	fmt.Println(r)
}
