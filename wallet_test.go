package go_coin_sol

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/pefish/go-coin-sol/program/pumpfun"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	type_ "github.com/pefish/go-coin-sol/type"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

var WalletInstance *Wallet

func init() {
	instance, err := New(context.Background(), &i_logger.DefaultLogger, "", "")
	if err != nil {
		panic(err)
	}
	WalletInstance = instance
}

func TestWallet_SendTx(t *testing.T) {
	return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("ESxwAtD82mHgPDvQ2D1j1EEi5UFYXGRFB4MdWLq8pump")
	data, err := pumpfun.GetBondingCurveData(WalletInstance.rpcClient, &tokenAddress, nil)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := pumpfun.GetSwapInstructions(
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		"6000",
		true,
		data.VirtualSolReserves,
		data.VirtualTokenReserves,
		50,
	)
	go_test_.Equal(t, nil, err)
	meta, tx, err := WalletInstance.SendTx(
		context.Background(),
		privObj,
		nil,
		swapInstructions,
		10000000,
		pumpfun_constant.Pumpfun_Buy_Unit_Limit,
		false,
	)
	go_test_.Equal(t, nil, err)
	swapResult, err := pumpfun.ParseSwapTx(meta, tx)
	go_test_.Equal(t, nil, err)
	fmt.Println(swapResult)
}

func TestWallet_NewAddress(t *testing.T) {
	address, priv := WalletInstance.NewAddress()
	fmt.Println(address, priv)
}
