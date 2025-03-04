package main

import (
	"context"
	"log"
	"math"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/joho/godotenv"
	go_coin_sol "github.com/pefish/go-coin-sol"
	"github.com/pefish/go-coin-sol/constant"
	t_logger "github.com/pefish/go-interface/t-logger"
	go_logger "github.com/pefish/go-logger"
)

func main() {
	err := do()
	if err != nil {
		log.Fatal(err)
	}
}

func do() error {
	envMap, _ := godotenv.Read("./.env")
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	wallet := go_coin_sol.New(
		go_logger.NewLogger(t_logger.Level_DEBUG),
		os.Getenv("NODE_HTTPS"),
		os.Getenv("NODE_WSS"),
	)
	privObj := solana.MustPrivateKeyFromBase58(os.Getenv("PRIV"))

	for {
		// time.Sleep(30 * time.Second)
		// continue
		instructions, err := wallet.TransferSOL(
			privObj.PublicKey(),
			solana.MustPublicKeyFromBase58("5BnsHy3CV2SjefwMPQ4pwQPVmigxA8R7gUZypRNsZqxp"),
			10000,
		)
		if err != nil {
			return err
		}
		// return
		_, err = wallet.SendTxByJitoV2(
			context.Background(),
			privObj,
			nil,
			nil,
			instructions,
			0,
			0,
			[]string{
				"https://tokyo.mainnet.block-engine.jito.wtf",
				"https://mainnet.block-engine.jito.wtf",
			},
			uint64(0.00002*math.Pow(10, constant.SOL_Decimals)),
			solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
		)
		if err != nil {
			return err
		}
		// wallet.WSClient().Close()
		time.Sleep(40 * time.Second)
	}
	return nil
}
