package sol_fi

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

func TestParseSwapTxByParsedTx(t *testing.T) {
	// return
	// 5fboj3wb4qdwTYSEQiFLBp6YhBpRXQ2T33dkP2LTrh4F23mUEyoRcL93ezHek8bEkJ7NFdFMrs4LoDHYDvPymTgq
	// 3pcBgtqfdmirMk6UJpdjxro7rv3QtcqsP5RiMpCKoV8w3Jo6HTDAqWbUdQEsa337ANsyrDRjyanw7EUqfVzAiyef
	// 3pFWV4dZBkUVXwj3WkXm7DasfCKXvxNjnahGJzMX6Ur6WBgowhEFfnUMXiQVNJXkKvHbnMkk6SqZWPy8UcKUYB84
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("3pFWV4dZBkUVXwj3WkXm7DasfCKXvxNjnahGJzMX6Ur6WBgowhEFfnUMXiQVNJXkKvHbnMkk6SqZWPy8UcKUYB84"),
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
<PairAddress: %s>
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
			swap.PairAddress,
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
