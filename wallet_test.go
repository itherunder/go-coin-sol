package go_coin_sol

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	constant "github.com/pefish/go-coin-sol/constant"
	"github.com/pefish/go-coin-sol/program/jupiter"
	"github.com/pefish/go-coin-sol/program/pumpfun"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	pumpfun_instruction "github.com/pefish/go-coin-sol/program/pumpfun/instruction"
	raydium_amm "github.com/pefish/go-coin-sol/program/raydium-amm"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium-amm/type"
	raydium_clmm "github.com/pefish/go-coin-sol/program/raydium-clmm"
	raydium_clmm_type "github.com/pefish/go-coin-sol/program/raydium-clmm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_test_ "github.com/pefish/go-test"
)

var WalletInstance *Wallet

func init() {
	url := rpc.MainNetBeta_RPC
	envUrl := os.Getenv("URL")
	if envUrl != "" {
		url = envUrl
	}
	WalletInstance = New(
		&i_logger.DefaultLogger,
		url,
		"",
	)
}

func TestWallet_SwapPumpfun(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("CcZJFmUJ95vX4Ae4g2SCjQzT8hGqFsQdPi5WeD9Qpump")
	data, err := pumpfun.GetBondingCurveData(rpc.MainNetBeta, WalletInstance.rpcClient, &tokenAddress, nil)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := pumpfun.GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Buy,
		tokenAddress,
		uint64(300*math.Pow(10, pumpfun_constant.Pumpfun_Token_Decimals)),
		true,
		data.VirtualSolReserveWithDecimals,
		data.VirtualTokenReserveWithDecimals,
		50,
	)
	go_test_.Equal(t, nil, err)
	r, err := WalletInstance.SendTx(
		context.Background(),
		privObj,
		nil,
		nil,
		swapInstructions,
		0,
		pumpfun_constant.Pumpfun_Buy_Unit_Limit,
		false,
		nil,
	)
	go_test_.Equal(t, nil, err)
	swapResult, err := pumpfun.ParseSwapTxByParsedTx(rpc.MainNetBeta, r.Meta, r.Transaction)
	go_test_.Equal(t, nil, err)
	fmt.Println(swapResult)
}

func TestWallet_NewAddress(t *testing.T) {
	address, priv := WalletInstance.NewAddress()
	fmt.Println(address, priv)
}

func TestWallet_SwapRaydiumAmm(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump")
	raydiumSwapKeys := raydium_type_.SwapKeys{
		AmmAddress: solana.MustPublicKeyFromBase58("4AZRPNEfCJ7iw28rJu5aUyeQhYcvdcNm8cswyL51AY9i"),
		PCMint:     tokenAddress,
		CoinMint:   solana.SolMint,
		Vaults: map[solana.PublicKey]solana.PublicKey{
			solana.SolMint: solana.MustPublicKeyFromBase58("AEwsZFbKVzf2MqADSHHhwqyWmTWYzruTG1HkMw8Mjq5"),
			tokenAddress:   solana.MustPublicKeyFromBase58("2zxMeSRkYa462Zo7v5K7kFKtvpRC4MpvuC1HwA88sCR3"),
		},
	}
	solAmount, tokenAmount, err := util.GetReserves(
		WalletInstance.rpcClient,
		raydiumSwapKeys.Vaults[raydiumSwapKeys.CoinMint],
		raydiumSwapKeys.Vaults[raydiumSwapKeys.PCMint],
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := raydium_amm.GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Sell,
		tokenAddress,
		uint64(4*math.Pow(10, pumpfun_constant.Pumpfun_Token_Decimals)),
		raydiumSwapKeys,
		true,
		solAmount.AmountWithDecimals,
		tokenAmount.AmountWithDecimals,
		50,
	)
	go_test_.Equal(t, nil, err)
	_, err = WalletInstance.SendTx(
		context.Background(),
		privObj,
		nil,
		nil,
		swapInstructions,
		1000000,
		raydium_constant.Raydium_Buy_Unit_Limit,
		false,
		nil,
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_SwapJupiter_Buy_Sell(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump")
	// amount := uint64(1 * math.Pow(10, pumpfun_constant.Pumpfun_Token_Decimals))

	swapInstructions, err := jupiter.GetSwapInstructions(
		&i_logger.DefaultLogger,
		privObj.PublicKey(),
		tokenAddress,
		&jupiter.QuoteType{
			SwapMode:             "ExactIn",
			InputMint:            solana.SolMint.String(),
			InAmount:             "859370",
			OutputMint:           tokenAddress.String(),
			OutAmount:            "849371",
			OtherAmountThreshold: "849371",
			SlippageBps:          500,
			PriceImpactPct:       "0",
			RoutePlan: []jupiter.RoutePlanType{
				{
					SwapInfo: jupiter.SwapInfoType{
						AmmKey:     "4AZRPNEfCJ7iw28rJu5aUyeQhYcvdcNm8cswyL51AY9i",
						Label:      "Raydium",
						InputMint:  solana.SolMint.String(),
						OutputMint: tokenAddress.String(),
						InAmount:   "859370",
						OutAmount:  "913700482",
						FeeAmount:  "1000",
						FeeMint:    solana.SolMint.String(),
					},
					Percent: 100,
				},
				{
					SwapInfo: jupiter.SwapInfoType{
						AmmKey:     "8oT91ooChsr7aHTHha9oJxKTYwUhZ75tjJ6bhtiggG5Y",
						Label:      "Raydium CLMM",
						InputMint:  tokenAddress.String(),
						OutputMint: solana.SolMint.String(),
						InAmount:   "913700482",
						OutAmount:  "849371",
						FeeAmount:  "1000",
						FeeMint:    tokenAddress.String(),
					},
					Percent: 100,
				},
			},
		},
		false,
	)
	go_test_.Equal(t, nil, err)
	_, err = WalletInstance.SendTxByJito(
		context.Background(),
		privObj,
		nil,
		nil,
		swapInstructions,
		0,
		0,
		[]string{
			"https://tokyo.mainnet.block-engine.jito.wtf",
			"https://mainnet.block-engine.jito.wtf",
		},
		uint64(0.0001*math.Pow(10, constant.SOL_Decimals)),
		solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_SwapJupiter_Buy(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump")

	quote, err := jupiter.GetQuote(
		&i_logger.DefaultLogger,
		type_.SwapType_Buy,
		tokenAddress,
		uint64(1*math.Pow(10, pumpfun_constant.Pumpfun_Token_Decimals)),
		50,
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := jupiter.GetSwapInstructions(
		&i_logger.DefaultLogger,
		privObj.PublicKey(),
		tokenAddress,
		quote,
		false,
	)
	go_test_.Equal(t, nil, err)
	_, err = WalletInstance.SendTxByJito(
		context.Background(),
		privObj,
		nil,
		nil,
		swapInstructions,
		0,
		0,
		[]string{
			"https://tokyo.mainnet.block-engine.jito.wtf",
			"https://mainnet.block-engine.jito.wtf",
		},
		uint64(0.00002*math.Pow(10, constant.SOL_Decimals)),
		solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_SwapJupiter_Sell(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("2qEHjDLDLbuBgRYvsxhc5D6uDWAivNFZGan56P1tpump")

	quote, err := jupiter.GetQuote(
		&i_logger.DefaultLogger,
		type_.SwapType_Sell,
		tokenAddress,
		uint64(1*math.Pow(10, pumpfun_constant.Pumpfun_Token_Decimals)),
		500,
	)
	go_test_.Equal(t, nil, err)

	swapInstructions, err := jupiter.GetSwapInstructions(
		&i_logger.DefaultLogger,
		privObj.PublicKey(),
		tokenAddress,
		quote,
		true,
	)
	go_test_.Equal(t, nil, err)
	_, err = WalletInstance.SendTxByJito(
		context.Background(),
		privObj,
		nil,
		nil,
		swapInstructions,
		0,
		0,
		[]string{
			"https://tokyo.mainnet.block-engine.jito.wtf",
			"https://mainnet.block-engine.jito.wtf",
		},
		uint64(0.00002*math.Pow(10, constant.SOL_Decimals)),
		solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_SwapRaydiumClmm_Buy(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddress := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	tokenDecimals := 6
	swapKeys := raydium_clmm_type.SwapV2Keys{
		PairAddress: solana.MustPublicKeyFromBase58("8sLbNZoA1cfnvMJLPfp98ZLAnFSYCFApfJKMbiXNLwxj"),
		Vaults: map[solana.PublicKey]solana.PublicKey{
			solana.SolMint: solana.MustPublicKeyFromBase58("6P4tvbzRY6Bh3MiWDHuLqyHywovsRwRpfskPvyeSoHsz"),
			tokenAddress:   solana.MustPublicKeyFromBase58("6mK4Pxs6GhwnessH7CvPivqDYauiHZmAdbEFDpXFk9zt"),
		},
		ObservationState: solana.MustPublicKeyFromBase58("3MsJXVvievxAbsMsaT6TS4i6oMitD9jazucuq3X234tC"),
		ExBitmapAccount:  solana.MustPublicKeyFromBase58("DoPuiZfJu7sypqwR4eiU7C5TMcmmiFoU4HaF5SoD8mRy"),
		RemainAccounts: []solana.PublicKey{
			solana.MustPublicKeyFromBase58("G9exbQ2QKCkZTviZvUZqG2NQco4cnAwuNCeqEnm18ta6"),
			solana.MustPublicKeyFromBase58("DRN2L8Tt6Fsz7CCtuAArRjSuPjXKz2tJhnvt6YrDoXFo"),
			solana.MustPublicKeyFromBase58("7NeQvZ8KrvU3Pbqb4vM5NWmajAVyLMsKyi7NASucybdY"),
		},
	}
	solAmount, tokenAmount, err := util.GetReserves(
		WalletInstance.rpcClient,
		swapKeys.Vaults[solana.SolMint],
		swapKeys.Vaults[tokenAddress],
	)
	go_test_.Equal(t, nil, err)
	fmt.Println("solAmount", solAmount.AmountWithDecimals)
	fmt.Println("tokenAmount", tokenAmount.AmountWithDecimals)

	swapInstructions, err := raydium_clmm.GetSwapInstructions(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		type_.SwapType_Buy,
		tokenAddress,
		uint64(1*math.Pow(10, float64(tokenDecimals))),
		swapKeys,
		true,
		solAmount.AmountWithDecimals,
		tokenAmount.AmountWithDecimals,
		500,
	)
	go_test_.Equal(t, nil, err)
	recent, err := WalletInstance.rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	go_test_.Equal(t, nil, err)
	latestBlockhash := &recent.Value.Blockhash
	_, err = WalletInstance.SendTxByJito(
		context.Background(),
		privObj,
		nil,
		latestBlockhash,
		swapInstructions,
		0,
		0,
		[]string{
			"https://tokyo.mainnet.block-engine.jito.wtf",
			"https://mainnet.block-engine.jito.wtf",
		},
		uint64(0.00002*math.Pow(10, constant.SOL_Decimals)),
		solana.MustPublicKeyFromBase58("DttWaMuVvTiduZRnguLF7jNxTgiMBZ1hyAumKUiL2KRL"),
	)
	go_test_.Equal(t, nil, err)
}

func TestWallet_DecodeProgramDataInLog(t *testing.T) {
	// EBd6RndiETniYBUREhYbYd9Ur1NrgdrDAnZfyAWqqYgLp4kNBY3UBghUAerddDqquX5dj3RUuy7s24cn3VH6xcG
	b, err := base64.RawStdEncoding.DecodeString("vdt/007mYe4+dQga6O8cSVy9IJsAPfc8BNC1hUfe1jIxkxCQ/2eXr48Pm1kAAAAA4n0OuJEuAAAAtBUSa0Qo7DENjrIxwgVuIRrUtrNJSCYNFytWRQseZyoCWnJnAAAAAAevI/wGAAAA1jczRuPPAwAHAwAAAAAAANafIPpR0QIA")
	go_test_.Equal(t, nil, err)
	fmt.Println(hex.EncodeToString(b))
	// WalletInstance.DecodeProgramDataInLog(
	// 	"",
	// 	&a,
	// )
}

func TestWallet_GetJitoTipInfo(t *testing.T) {
	// return
	info, err := WalletInstance.GetJitoTipInfo()
	go_test_.Equal(t, nil, err)
	fmt.Println(info.EMALandedTips50thPercentile)
}

func TestWallet_TokenBalance(t *testing.T) {
	info, err := WalletInstance.TokenBalance(
		solana.MustPublicKeyFromBase58("Gr1KhnM4sjzwHnnLbVPMVgQcv2AXwaP7m2U8k3PKcNXz"),
		solana.MustPublicKeyFromBase58("EJJ1EdGLAyd97AMqF3xBT4HT8uvBavcR2US5eM7vVsF9"),
	)
	go_test_.Equal(t, nil, err)
	fmt.Println(info.AmountWithDecimals)
}

func TestWallet_ParseRaydiumAddLiqRayLog(t *testing.T) {
	b, err := base64.StdEncoding.DecodeString("AAAAAAAAAAAABglkAAAAAAAAAICWmAAAAAAAAAgBqSy8AAA2IBZlEgAAAGoftr4VXUHxXQSto4B86dMJI9cALHyyo/r6yWIFQ4d9")
	go_test_.Equal(t, nil, err)
	var logObj struct {
		LogType      uint8            `json:"log_type"`
		Timestamp    uint64           `json:"time"`
		PcDecimals   uint8            `json:"pc_decimals"`
		CoinDecimals uint8            `json:"coin_decimals"`
		PcLotSize    uint64           `json:"pc_lot_size"`
		CoinLotSize  uint64           `json:"coin_lot_size"`
		PcAmount     uint64           `json:"pc_amount"`
		CoinAmount   uint64           `json:"coin_amount"`
		Market       solana.PublicKey `json:"market"`
	}
	err = bin.NewBorshDecoder(b).Decode(&logObj)
	go_test_.Equal(t, nil, err)
	fmt.Println(logObj)
}

func TestGetTokenData(t *testing.T) {
	r, err := WalletInstance.GetTokenData(solana.MustPublicKeyFromBase58("PELGx59WwJXY83tbr85XGyVoHU7MHTJB8wP2PRiLmM9"))
	go_test_.Equal(t, nil, err)
	fmt.Println(r)
}

func TestCreatePumpfunToken(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	tokenAddressWallet := solana.NewWallet()

	bondingCurveAddress, err := pumpfun.DeriveBondingCurveAddress(rpc.MainNetBeta, tokenAddressWallet.PublicKey())
	go_test_.Equal(t, nil, err)
	createInstruction, err := pumpfun_instruction.NewCreateInstruction(
		rpc.MainNetBeta,
		privObj.PublicKey(),
		tokenAddressWallet.PublicKey(),
		bondingCurveAddress,
		"haha",
		"HAHA",
		"https://ipfs.io/ipfs/QmdxfAvJ8gZr3XcdUnB3XwMFYcLsoTLE2efsLvsrwUBx7u",
	)
	go_test_.Equal(t, nil, err)
	swapInstructions, err := pumpfun.GetSwapInstructions(
		rpc.DevNet,
		privObj.PublicKey(),
		type_.SwapType_Buy,
		tokenAddressWallet.PublicKey(),
		10000000000,
		false,
		30000000000,
		1000000000000000,
		50,
	)
	go_test_.Equal(t, nil, err)
	recent, err := WalletInstance.rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	go_test_.Equal(t, nil, err)
	latestBlockhash := &recent.Value.Blockhash
	instructions := []solana.Instruction{
		createInstruction,
	}
	instructions = append(instructions, swapInstructions...)
	_, err = WalletInstance.SendTx(
		context.Background(),
		privObj,
		map[solana.PublicKey]*solana.PrivateKey{
			privObj.PublicKey():            &privObj,
			tokenAddressWallet.PublicKey(): &tokenAddressWallet.PrivateKey,
		},
		latestBlockhash,
		instructions,
		0,
		0,
		true,
		nil,
	)
	go_test_.Equal(t, nil, err)
}

func TestGetDestroyTokenAccountInstructions(t *testing.T) {
	// return
	privObj, err := solana.PrivateKeyFromBase58(os.Getenv("PRIV"))
	go_test_.Equal(t, nil, err)
	instructions, err := WalletInstance.GetDestroyTokenAccountInstructions(
		solana.MustPublicKeyFromBase58("CkvFJefgLGMAtFeCRU7cbZJgrwj5vBUGtDv3mcfnanzK"),
		[]solana.PublicKey{
			solana.MustPublicKeyFromBase58("66TSDhxGEBxqkc4sNAxiVGiCTxtdJVxidQa3PxuSAZA1"),
			solana.MustPublicKeyFromBase58("6D2D5bMwrk2ayuCi66Prgob3q7KGZYVy6q2UEqqgcuQh"),
			solana.MustPublicKeyFromBase58("HTdEJzCjkpULdZysPVm7wU9VJqfA6N4ZqntM5Lf1qVXU"),
			solana.MustPublicKeyFromBase58("2PCaMNLBKTj7jg9fXQmk4tCam1uGLJFwnCYRK47gGRDb"),
			solana.MustPublicKeyFromBase58("5qADJuYPwNqgR1DWUwMM479iZe5Vq7P4RyMqkR93kLtj"),
			solana.MustPublicKeyFromBase58("211HHjYaPicnwkxPb64ypPJrNXv75fApZaXzddSwSFTs"),
			solana.MustPublicKeyFromBase58("j9RMNNZxze8noZi3oayfhxfQwxLfDsCPyhcMcqamDZq"),
			solana.MustPublicKeyFromBase58("E4MC9kjz8zVDK5Xo7S5RLETzh3SKX7PhaqJxh9Xoxgqd"),
		},
	)
	go_test_.Equal(t, nil, err)
	_, err = WalletInstance.SendTxByJito(
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
	go_test_.Equal(t, nil, err)
}
