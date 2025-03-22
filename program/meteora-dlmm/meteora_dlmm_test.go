package meteora_dlmm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/itherunder/go-coin-sol/constant"
	meteora_dlmm_type "github.com/itherunder/go-coin-sol/program/meteora-dlmm/type"
	go_format "github.com/itherunder/go-format"
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

func TestParseSwapTxByParsedTx(t *testing.T) {
	// return
	// 56Nx6pvvQv6e3dwp58SK19fiTW8ne9ATddCFHvD73q2FrPyZFUpufhsxBQpQYMfbNz7o2wmBeFUmahstMi3p9wtm
	// 4czxuc1Bm3KimGJEx7QZTx92S8b475nBiapVouN1NtTdNCs1uYYqG5NexXmxVTocxXgFdpvDZJeuAHRJFYJCMCwN
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("4czxuc1Bm3KimGJEx7QZTx92S8b475nBiapVouN1NtTdNCs1uYYqG5NexXmxVTocxXgFdpvDZJeuAHRJFYJCMCwN"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(rpc.MainNetBeta, getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	for _, swap := range r.Swaps {
		parsedKeys := make([]rpc.ParsedMessageAccount, 0)
		for _, a := range swap.Keys {
			for _, b := range swap.AllKeys {
				if a.Equals(b.PublicKey) {
					parsedKeys = append(parsedKeys, b)
					break
				}
			}
		}

		extraDatas := swap.ExtraDatas.(*meteora_dlmm_type.ExtraDatasType)
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<OutputAddress: %s>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<InputVault: %s>
<OutputVault: %s>
<Keys: %s>
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
			go_format.ToString(parsedKeys),
			extraDatas.ReserveInputWithDecimals,
			extraDatas.ReserveOutputWithDecimals,
		)
	}
}
