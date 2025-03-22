package sol_fi

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/constant"
	sol_fi_type "github.com/itherunder/go-coin-sol/program/sol-fi/type"
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

func TestDecodeSwapInstructionData(t *testing.T) {
	// return
	b, err := hex.DecodeString("073c5b702200000000000000000000000001")
	go_test_.Equal(t, nil, err)
	var r struct {
		Discriminator uint8
		AmountIn      uint64
		Unknown1      uint64
		BToA          bool
	}
	err = bin.NewBorshDecoder(b).Decode(&r)
	go_test_.Equal(t, nil, err)
	fmt.Println(r.Discriminator, r.AmountIn, r.BToA)
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
		extraDatas := swap.ExtraDatas.(*sol_fi_type.ExtraDatasType)
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
			extraDatas.ReserveInputWithDecimals,
			extraDatas.ReserveOutputWithDecimals,
		)
	}

}
