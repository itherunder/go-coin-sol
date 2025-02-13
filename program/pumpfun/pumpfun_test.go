package pumpfun

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	type_ "github.com/pefish/go-coin-sol/program/pumpfun/type"
	i_logger "github.com/pefish/go-interface/i-logger"
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
		solana.MustSignatureFromBase58("2wKzg8JRs6SZp655aL31dVkmaAqaaHEtt8GYZRYHhuUsSXDac7Nq8T1j7a2oSFBpuJPnztXY6utt4aGxL3WAz22B"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseSwapTxByParsedTx(getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	for _, swap := range r.Swaps {
		extraDatas := swap.ExtraDatas.(*type_.ExtraDatasType)
		fmt.Printf(
			`
<UserAddress: %s>
<InputAddress: %s>
<OutputAddress: %s>
<InputAmountWithDecimals: %d>
<OutputAmountWithDecimals: %d>
<ReserveSOLAmountWithDecimals: %d>
<ReserveTokenAmountWithDecimals: %d>
<Timestamp: %d>
`,
			swap.UserAddress,
			swap.InputAddress,
			swap.OutputAddress,
			swap.InputAmountWithDecimals,
			swap.OutputAmountWithDecimals,
			extraDatas.ReserveSOLAmountWithDecimals,
			extraDatas.ReserveTokenAmountWithDecimals,
			extraDatas.Timestamp,
		)
	}
}

func TestURIInfo(t *testing.T) {
	r, err := URIInfo(&i_logger.DefaultLogger, "https://ipfs.io/ipfs/QmVSKrX4XxUgHMCp2wnmE3VmnK3fCyZsSqFiVGQZowiM1c")
	go_test_.Equal(t, nil, err)
	fmt.Println("aa", r.Twitter)
}

func TestGetBondingCurveData(t *testing.T) {
	tokenAddressObj := solana.MustPublicKeyFromBase58("7PAaQ8UxYYPksnyxcKFP44Pm4FwFCix4ammGf5P3bK79")
	r, err := GetBondingCurveData(client, &tokenAddressObj, nil)
	go_test_.Equal(t, nil, err)
	fmt.Println(r)
}

func TestGenerateTokenURI(t *testing.T) {
	// return
	r, err := GenerateTokenURI(&GenerateTokenURIDataType{
		Name:        "testcoin",
		Symbol:      "TEST",
		Description: "test test.",
		File:        nil,
		Twitter:     "https://x.com",
		Website:     "https://x.com",
		Telegram:    "https://tg.com",
	})
	go_test_.Equal(t, nil, err)
	fmt.Printf("%#v\n", r)
}

func TestGenePumpfunWallet(t *testing.T) {
	// return
	r, err := GenePumpfunWallet(2 * time.Minute)
	go_test_.Equal(t, nil, err)
	fmt.Println(r.PublicKey().String())
}

func TestParseRemoveLiqTxByParsedTx(t *testing.T) {
	// return
	getTransactionResult, err := client.GetParsedTransaction(
		context.TODO(),
		solana.MustSignatureFromBase58("Dw9KGZqJ9CAB5PWe89r9VoDTfW97Hw49EN6GjLZ5W4RLrzx8i2sLny4uBqXHqQkZNZt8FCGcoTonqrj6JrkZkDm"),
		&rpc.GetParsedTransactionOpts{
			Commitment:                     rpc.CommitmentConfirmed,
			MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
		},
	)
	go_test_.Equal(t, nil, err)
	r, err := ParseRemoveLiqTxByParsedTx(getTransactionResult.Meta, getTransactionResult.Transaction)
	go_test_.Equal(t, nil, err)
	go_test_.Equal(t, false, r == nil)
	fmt.Printf(
		"[RemoveLiq] <%s> <BondingCurveAddress: %s>\n",
		r.TokenAddress,
		r.BondingCurveAddress.String(),
	)
}
