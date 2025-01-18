package go_coin_sol

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	constant "github.com/pefish/go-coin-sol/constant"
	associated_token_account "github.com/pefish/go-coin-sol/program/associated-token-account"
	type_ "github.com/pefish/go-coin-sol/type"
	go_format "github.com/pefish/go-format"
	go_http "github.com/pefish/go-http"
	i_logger "github.com/pefish/go-interface/i-logger"
	go_time "github.com/pefish/go-time"
	"github.com/pkg/errors"
)

type Wallet struct {
	logger    i_logger.ILogger
	rpcClient *rpc.Client
	wssUrl    string
}

func New(
	ctx context.Context,
	logger i_logger.ILogger,
	httpsUrl string,
	wssUrl string,
) (*Wallet, error) {
	if httpsUrl == "" {
		httpsUrl = rpc.MainNetBeta_RPC
	}
	if wssUrl == "" {
		wssUrl = rpc.MainNetBeta_WS
	}
	rpcClient := rpc.New(httpsUrl)
	return &Wallet{
		logger:    logger,
		rpcClient: rpcClient,
		wssUrl:    wssUrl,
	}, nil
}

func (t *Wallet) RPCClient() *rpc.Client {
	return t.rpcClient
}

func (t *Wallet) NewWSClient(ctx context.Context) (*ws.Client, error) {
	wsClient, err := ws.Connect(ctx, t.wssUrl)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return wsClient, nil
}

func (t *Wallet) NewAddress() (address_ string, priv_ string) {
	account := solana.NewWallet()
	return account.PublicKey().String(), account.PrivateKey.String()
}

func (t *Wallet) TokenBalance(
	address solana.PublicKey,
	tokenAddress solana.PublicKey,
) (*type_.TokenAmountInfo, error) {
	userTokenAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		address,
		tokenAddress,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	data, err := associated_token_account.GetAssociatedTokenAccountData(
		t.rpcClient,
		userTokenAssociatedAccount,
	)
	if err != nil {
		if err.Error() == "not found" {
			return &type_.TokenAmountInfo{
				AmountWithDecimals: 0,
				Decimals:           0,
			}, nil
		}
		return nil, errors.Wrap(err, "")
	}
	if data == nil {
		return &type_.TokenAmountInfo{
			AmountWithDecimals: 0,
			Decimals:           0,
		}, nil
	}
	amountWithDecimals, err := strconv.ParseUint(data.Parsed.Info.TokenAmount.Amount, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &type_.TokenAmountInfo{
		AmountWithDecimals: amountWithDecimals,
		Decimals:           data.Parsed.Info.TokenAmount.Decimals,
	}, nil
}

func (t *Wallet) SendTx(
	ctx context.Context,
	privObj solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
	skipPreflight bool,
	urls []string,
) (
	meta_ *rpc.TransactionMeta,
	tx_ *solana.Transaction,
	timestamp_ uint64,
	err_ error,
) {
	tx, err := t.BuildTx(privObj, latestBlockhash, instructions, unitPrice, unitLimit)
	if err != nil {
		return nil, nil, 0, err
	}
	t.logger.InfoF("交易构建成功 <%d>。<%s>", go_time.CurrentTimestamp(), tx.Signatures[0].String())

	meta, timestamp, err := t.SendAndConfirmTransaction(ctx, tx, skipPreflight, urls)
	if err != nil {
		return nil, nil, 0, err
	}

	return meta, tx, timestamp, nil
}

func (t *Wallet) SendTxByJito(
	ctx context.Context,
	privObj solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
	jitoTipAmountWithDecimals uint64,
	jitoAccount solana.PublicKey,
) (
	meta_ *rpc.TransactionMeta,
	tx_ *solana.Transaction,
	timestamp_ uint64,
	err_ error,
) {

	instructions = append(
		instructions,
		system.
			NewTransferInstructionBuilder().
			SetFundingAccount(privObj.PublicKey()).
			SetRecipientAccount(jitoAccount).
			SetLamports(jitoTipAmountWithDecimals).
			Build(),
	)

	tx, err := t.BuildTx(privObj, latestBlockhash, instructions, unitPrice, unitLimit)
	if err != nil {
		return nil, nil, 0, err
	}
	t.logger.InfoF("交易构建成功 <%d>。<%s>", go_time.CurrentTimestamp(), tx.Signatures[0].String())

	meta, timestamp, err := t.SendByJitoAndConfirmTransaction(ctx, tx)
	if err != nil {
		return nil, nil, 0, err
	}

	return meta, tx, timestamp, nil
}

func (t *Wallet) BuildTx(
	privObj solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
) (*solana.Transaction, error) {
	userAddress := privObj.PublicKey()

	feeInstructions := make([]solana.Instruction, 0)
	if unitPrice != 0 {
		feeInstructions = append(feeInstructions, computebudget.NewSetComputeUnitLimitInstruction(uint32(unitLimit)).Build())
		feeInstructions = append(feeInstructions, computebudget.NewSetComputeUnitPriceInstruction(unitPrice).Build())
	}
	if latestBlockhash == nil {
		recent, err := t.rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		latestBlockhash = &recent.Value.Blockhash
	}
	tx, err := solana.NewTransaction(
		append(
			feeInstructions,
			instructions...,
		),
		*latestBlockhash,
		solana.TransactionPayer(userAddress),
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	tx.Message.SetVersion(solana.MessageVersionV0)

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key.Equals(userAddress) {
				return &privObj
			}
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return tx, nil
}

type JitoTipInfo struct {
	Time                        string  `json:"time"`
	LandedTips25thPercentile    float64 `json:"landed_tips_25th_percentile"`
	LandedTips50thPercentile    float64 `json:"landed_tips_50th_percentile"`
	LandedTips75thPercentile    float64 `json:"landed_tips_75th_percentile"`
	LandedTips95thPercentile    float64 `json:"landed_tips_95th_percentile"`
	LandedTips99thPercentile    float64 `json:"landed_tips_99th_percentile"`
	EMALandedTips50thPercentile float64 `json:"ema_landed_tips_50th_percentile"`
}

func (t *Wallet) GetJitoTipInfo() (*JitoTipInfo, error) {
	var httpResult []*JitoTipInfo
	_, _, err := go_http.NewHttpRequester(
		go_http.WithLogger(t.logger),
		go_http.WithTimeout(5*time.Second),
	).GetForStruct(
		&go_http.RequestParams{
			Url: "https://bundles.jito.wtf/api/v1/bundles/tip_floor",
		},
		&httpResult,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return httpResult[0], nil
}

func (t *Wallet) SendByJitoAndConfirmTransaction(
	ctx context.Context,
	tx *solana.Transaction,
) (
	meta_ *rpc.TransactionMeta,
	timestamp_ uint64,
	err_ error,
) {
	rpcClient := rpc.New("https://tokyo.mainnet.block-engine.jito.wtf/api/v1/transactions")
	newCtx, _ := context.WithTimeout(ctx, 90*time.Second) // 150 个 slot 链上就会超时，每个 slot 是 400ms - 600ms，也就是 60-90s
	sendTimer := time.NewTimer(0)
	confirmTimer := time.NewTimer(time.Second)
	for {
		select {
		case <-sendTimer.C:
			_, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight: true,
			})
			if err != nil {
				if strings.Contains(err.Error(), "already processed transaction") {
					sendTimer.Stop()
					continue
				}
				t.logger.WarnF("交易发送失败. <%s>", err.Error())
				sendTimer.Reset(time.Second)
				continue
			}
			// t.logger.InfoF("交易已发送。<%s>", tx.Signatures[0].String())
			sendTimer.Reset(500 * time.Millisecond)
			continue
		case <-confirmTimer.C:
			getTransactionResult, err := t.rpcClient.GetTransaction(
				ctx,
				tx.Signatures[0],
				&rpc.GetTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil {
				t.logger.Debug(err)
				confirmTimer.Reset(2 * time.Second)
				continue
			}
			if getTransactionResult == nil {
				confirmTimer.Reset(2 * time.Second)
				continue
			}

			if getTransactionResult.Meta.Err != nil {
				t.logger.InfoF("交易已确认[执行失败] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
				return getTransactionResult.Meta, uint64(*getTransactionResult.BlockTime * 1000), errors.New(go_format.ToString(getTransactionResult.Meta.Err))
			}
			t.logger.InfoF("交易已确认[执行成功] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
			return getTransactionResult.Meta, uint64(*getTransactionResult.BlockTime * 1000), nil
		case <-newCtx.Done():
			return nil, 0, errors.New("确认超时")
		}
	}

}

func (t *Wallet) SendAndConfirmTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	skipPreflight bool,
	urls []string,
) (
	meta_ *rpc.TransactionMeta,
	timestamp_ uint64,
	err_ error,
) {
	for _, url := range urls {
		go func(url string) {
			rpc.New(url).SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight: skipPreflight,
			})
		}(url)
	}

	newCtx, _ := context.WithTimeout(ctx, 90*time.Second) // 150 个 slot 链上就会超时，每个 slot 是 400ms - 600ms，也就是 60-90s
	confirmTimer := time.NewTimer(0)
	for {
		select {
		case <-confirmTimer.C:
			_, err := t.rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight: skipPreflight,
			})
			if err != nil {
				if strings.Contains(err.Error(), "Blockhash not found") {
					confirmTimer.Reset(500 * time.Millisecond)
					continue
				}
				if strings.Contains(err.Error(), "Program failed to complete") ||
					strings.Contains(err.Error(), "custom program error") {
					return nil, 0, err
				}
				// t.logger.Error(err.Error())
			}
			getTransactionResult, err := t.rpcClient.GetTransaction(
				ctx,
				tx.Signatures[0],
				&rpc.GetTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil {
				t.logger.Debug(err)
				confirmTimer.Reset(2 * time.Second)
				continue
			}
			if getTransactionResult == nil {
				confirmTimer.Reset(2 * time.Second)
				continue
			}

			if getTransactionResult.Meta.Err != nil {
				t.logger.InfoF("交易已确认[执行失败] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
				return getTransactionResult.Meta, uint64(*getTransactionResult.BlockTime * 1000), errors.New(go_format.ToString(getTransactionResult.Meta.Err))
			}
			t.logger.InfoF("交易已确认[执行成功] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
			return getTransactionResult.Meta, uint64(*getTransactionResult.BlockTime * 1000), nil
		case <-newCtx.Done():
			return nil, 0, errors.New("确认超时")
		}
	}

}

type TokenDataType struct {
	Parsed struct {
		Info struct {
			Decimals        uint64 `json:"decimals"`
			FreezeAuthority string `json:"freezeAuthority"`
			IsInitialized   bool   `json:"isInitialized"`
			MintAuthority   string `json:"mintAuthority"`
			Supply          string `json:"supply"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   uint64 `json:"space"`
}

func (t *Wallet) GetTokenData(
	tokenAddress solana.PublicKey,
) (*TokenDataType, error) {
	var data TokenDataType
	r, err := t.rpcClient.GetAccountInfoWithOpts(
		context.Background(),
		tokenAddress,
		&rpc.GetAccountInfoOpts{
			Encoding: solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	err = json.Unmarshal(r.Value.Data.GetRawJSON(), &data)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return &data, nil
}
