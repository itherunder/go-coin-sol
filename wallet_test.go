package go_coin_sol

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/program/pumpfun"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	"github.com/pefish/go-coin-sol/program/raydium"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

var WalletInstance *Wallet

func init() {
	instance, err := New(context.Background(), &i_logger.DefaultLogger, rpc.MainNetBeta_RPC, "")
	if err != nil {
		panic(err)
	}
	WalletInstance = instance
}

func TestWallet_SwapPumpfun(t *testing.T) {
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
		"1000",
		true,
		data.VirtualSolReserves,
		data.VirtualTokenReserves,
		50,
	)
	go_test_.Equal(t, nil, err)
	meta, tx, _, err := WalletInstance.SendTx(
		context.Background(),
		privObj,
		nil,
		swapInstructions,
		1000000,
		pumpfun_constant.Pumpfun_Buy_Unit_Limit,
		false,
		nil,
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

func TestWallet_SwapRaydium(t *testing.T) {
	return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump")
	raydiumSwapKeys := raydium_type_.RaydiumSwapKeys{
		AmmAddress:                  solana.MustPublicKeyFromBase58("4AZRPNEfCJ7iw28rJu5aUyeQhYcvdcNm8cswyL51AY9i"),
		PoolCoinTokenAccountAddress: solana.MustPublicKeyFromBase58("AEwsZFbKVzf2MqADSHHhwqyWmTWYzruTG1HkMw8Mjq5"),
		PoolPcTokenAccountAddress:   solana.MustPublicKeyFromBase58("2zxMeSRkYa462Zo7v5K7kFKtvpRC4MpvuC1HwA88sCR3"),
	}
	solAmount, tokenAmount, err := raydium.GetReserves(
		WalletInstance.rpcClient,
		raydiumSwapKeys.PoolCoinTokenAccountAddress,
		raydiumSwapKeys.PoolPcTokenAccountAddress,
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := raydium.GetSwapInstructions(
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		type_.TokenAmountInfo{
			Amount:   "4",
			Decimals: pumpfun_constant.Pumpfun_Token_Decimals,
		},
		raydiumSwapKeys,
		true,
		solAmount,
		tokenAmount,
		50,
	)
	go_test_.Equal(t, nil, err)
	_, _, _, err = WalletInstance.SendTx(
		context.Background(),
		privObj,
		nil,
		swapInstructions,
		1000000,
		raydium_constant.Raydium_Buy_Unit_Limit,
		false,
		nil,
	)
	go_test_.Equal(t, nil, err)
}
