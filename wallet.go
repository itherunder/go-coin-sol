package go_coin_sol

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	jitorpc "github.com/jito-labs/jito-go-rpc"
	"github.com/mr-tron/base58"
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
	wsClient  *ws.Client
	wssUrl    string
}

func New(
	logger i_logger.ILogger,
	httpsUrl string,
	wssUrl string,
) *Wallet {
	if httpsUrl == "" {
		httpsUrl = rpc.MainNetBeta_RPC
	}
	if wssUrl == "" {
		wssUrl = rpc.MainNetBeta_WS
	}

	return &Wallet{
		logger:    logger,
		rpcClient: rpc.New(httpsUrl),
		wsClient:  ws.Connect(context.Background(), wssUrl),
		wssUrl:    wssUrl,
	}
}

func (t *Wallet) RPCClient() *rpc.Client {
	return t.rpcClient
}

func (t *Wallet) WSClient() *ws.Client {
	return t.wsClient
}

func (t *Wallet) NewWSClient(ctx context.Context, opt *ws.Options) *ws.Client {
	return ws.ConnectWithOptions(ctx, t.wssUrl, opt)
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
	signers map[solana.PublicKey]*solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
	skipPreflight bool,
	urls []string,
) (*rpc.GetParsedTransactionResult, error) {
	if len(instructions) == 0 {
		return nil, errors.Errorf("指令集为空")
	}

	tx, err := t.BuildTx(
		privObj,
		signers,
		latestBlockhash,
		instructions,
		unitPrice,
		unitLimit,
	)
	if err != nil {
		return nil, err
	}
	t.logger.InfoF("交易构建成功 <%d>。<%s>", go_time.CurrentTimestamp(), tx.Signatures[0].String())

	return t.SendAndConfirmTransaction(ctx, tx, skipPreflight, urls)
}

func (t *Wallet) SendTxByJito(
	ctx context.Context,
	privObj solana.PrivateKey,
	signers map[solana.PublicKey]*solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
	jitoUrls []string,
	jitoTipAmountWithDecimals uint64,
	jitoAccount solana.PublicKey,
) (*rpc.GetParsedTransactionResult, error) {
	if len(instructions) == 0 {
		return nil, errors.Errorf("指令集为空")
	}

	instructions = append(
		instructions,
		system.
			NewTransferInstructionBuilder().
			SetFundingAccount(privObj.PublicKey()).
			SetRecipientAccount(jitoAccount).
			SetLamports(jitoTipAmountWithDecimals).
			Build(),
	)

	tx, err := t.BuildTx(
		privObj,
		signers,
		latestBlockhash,
		instructions,
		unitPrice,
		unitLimit,
	)
	if err != nil {
		return nil, err
	}
	sendedTimestamp := go_time.CurrentTimestamp()
	t.logger.InfoF("交易构建成功 <timestamp: %d> <txid: %s>", sendedTimestamp, tx.Signatures[0].String())

	for _, jitoUrl := range jitoUrls {
		go func(jitoUrl string) {
			rpcClient := rpc.New(fmt.Sprintf("%s/api/v1/transactions", jitoUrl))
			_, err = rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight: true,
			})
			if err != nil {
				// t.logger.ErrorF("交易发送失败 <txid: %s> <jitoUrl: %s>. <%s>", tx.Signatures[0].String(), jitoUrl, err.Error())
				return
			}
			// t.logger.InfoF("交易发送成功 <txid: %s> <jitoUrl: %s>", tx.Signatures[0].String(), jitoUrl)
		}(jitoUrl)
	}

	newCtx, _ := context.WithTimeout(ctx, 90*time.Second) // 150 个 slot 链上就会超时，每个 slot 是 400ms - 600ms，也就是 60-90s
	confirmTimer := time.NewTimer(time.Second)

	var signatureSubscribeChan <-chan *ws.SignatureResult
	var sub *ws.SignatureSubscription
	go func() {
		sub, err = t.wsClient.SignatureSubscribe(
			tx.Signatures[0],
			rpc.CommitmentConfirmed,
		)
		if err != nil {
			t.logger.ErrorF("SignatureSubscribe failed. %s", err.Error())
			return
		}
		signatureSubscribeChan = sub.Response()
	}()
	defer func() {
		if sub != nil {
			sub.Unsubscribe()
		}
	}()

	var getTransactionResult *rpc.GetParsedTransactionResult
confirm:
	for {
		select {
		case _, ok := <-signatureSubscribeChan:
			if !ok {
				continue
			}
			getTransactionResult_, err := t.rpcClient.GetParsedTransaction(
				ctx,
				tx.Signatures[0],
				&rpc.GetParsedTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil || getTransactionResult_ == nil {
				continue
			}
			t.logger.InfoF("ws 检查到已确认")
			confirmTimer.Stop()
			getTransactionResult = getTransactionResult_
			break confirm
		case <-confirmTimer.C:
			getTransactionResult_, err := t.rpcClient.GetParsedTransaction(
				ctx,
				tx.Signatures[0],
				&rpc.GetParsedTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil || getTransactionResult_ == nil {
				go rpc.New(fmt.Sprintf("%s/api/v1/transactions", jitoUrls[0])).SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
					SkipPreflight: true,
				})
				// t.logger.InfoF("未确认...")
				confirmTimer.Reset(200 * time.Millisecond)
				continue
			}
			t.logger.InfoF("rpc 检查到已确认")
			confirmTimer.Stop()
			getTransactionResult = getTransactionResult_
			break confirm
		case <-newCtx.Done():
			return nil, errors.New("确认超时")
		}
	}
	if getTransactionResult.Meta.Err != nil {
		t.logger.InfoF(
			"交易已确认[执行失败] <timestamp: %d> <txid: %s> <历时: %ds>. <%s>",
			*getTransactionResult.BlockTime*1000,
			tx.Signatures[0].String(),
			(int64(*getTransactionResult.BlockTime*1000)-sendedTimestamp)/1000,
			getTransactionResult.Meta.Err,
		)
		return getTransactionResult, errors.Errorf("<txid: %s> <err: %s>", tx.Signatures[0], go_format.ToString(getTransactionResult.Meta.Err))
	}
	t.logger.InfoF(
		"交易已确认[执行成功] <timestamp: %d> <txid: %s> <历时: %ds>",
		*getTransactionResult.BlockTime*1000,
		tx.Signatures[0].String(),
		(int64(*getTransactionResult.BlockTime*1000)-sendedTimestamp)/1000,
	)
	return getTransactionResult, nil
}

func (t *Wallet) SendTxByJitoBundle(
	ctx context.Context,
	payFeePrivObj solana.PrivateKey,
	latestBlockhash *solana.Hash,
	txs []*solana.Transaction,
	jitoUrl string,
	jitoTipAmountWithDecimals uint64,
	jitoAccount solana.PublicKey,
) (
	timestamp_ uint64,
	err_ error,
) {
	sendFeeTx, err := t.BuildTx(
		payFeePrivObj,
		nil,
		latestBlockhash,
		[]solana.Instruction{
			system.
				NewTransferInstructionBuilder().
				SetFundingAccount(payFeePrivObj.PublicKey()).
				SetRecipientAccount(jitoAccount).
				SetLamports(jitoTipAmountWithDecimals).
				Build(),
		},
		0,
		0,
	)
	if err != nil {
		return 0, err
	}
	serializedSendFeeTx, _ := sendFeeTx.MarshalBinary()

	txIds := []string{
		sendFeeTx.Signatures[0].String(),
	}
	serializedTxs := []string{
		base58.Encode(serializedSendFeeTx),
	}
	for _, tx := range txs {
		serializedTx, _ := tx.MarshalBinary()
		serializedTxs = append(serializedTxs, base58.Encode(serializedTx))
		txIds = append(txIds, tx.Signatures[0].String())
	}

	jitoClient := jitorpc.NewJitoJsonRpcClient(fmt.Sprintf("%s/api/v1", jitoUrl), "")
	bundleIdRaw, err := jitoClient.SendBundle([][]string{
		serializedTxs,
	})
	if err != nil {
		return 0, errors.Wrap(err, "")
	}

	var bundleId string
	err = json.Unmarshal(bundleIdRaw, &bundleId)
	if err != nil {
		return 0, errors.Wrap(err, "")
	}
	sendedTimestamp := go_time.CurrentTimestamp()
	t.logger.InfoF("交易发送成功 <timestamp: %d> <txs: %s> <bundle_id: %s>", sendedTimestamp, go_format.ToString(txIds), bundleId)

	newCtx, _ := context.WithTimeout(ctx, 90*time.Second)
	confirmTimer := time.NewTimer(time.Second)
	for {
		select {
		case <-confirmTimer.C:
			statusResponse, err := jitoClient.GetBundleStatuses([]string{bundleId})
			if err != nil {
				confirmTimer.Reset(time.Second)
				continue
			}

			if len(statusResponse.Value) == 0 {
				confirmTimer.Reset(time.Second)
				continue
			}

			if statusResponse.Value[0].ConfirmationStatus != "confirmed" {
				confirmTimer.Reset(time.Second)
				continue
			}
			getTransactionResult, err := t.rpcClient.GetTransaction(
				ctx,
				txs[len(txs)-1].Signatures[0],
				&rpc.GetTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil {
				return 0, errors.Wrap(err, "")
			}
			if getTransactionResult == nil {
				return 0, errors.Errorf("<txid: %s> not found.", txs[len(txs)-1].Signatures[0])
			}
			if statusResponse.Value[0].Err.Ok != nil {
				t.logger.InfoF(
					"交易已确认[执行失败] <timestamp: %d> <txs: %s> <历时: %ds>. <%v>",
					*getTransactionResult.BlockTime*1000,
					go_format.ToString(txIds),
					(int64(*getTransactionResult.BlockTime*1000)-sendedTimestamp)/1000,
					statusResponse.Value[0].Err.Ok,
				)
				return uint64(*getTransactionResult.BlockTime * 1000), errors.Errorf("<txid: %s> <err: %s>", go_format.ToString(txIds), go_format.ToString(statusResponse.Value[0].Err.Ok))
			}
			t.logger.InfoF(
				"交易已确认[执行成功] <timestamp: %d> <txs: %s> <历时: %ds>",
				*getTransactionResult.BlockTime*1000,
				go_format.ToString(txIds),
				(int64(*getTransactionResult.BlockTime*1000)-sendedTimestamp)/1000,
			)
			return uint64(*getTransactionResult.BlockTime * 1000), nil

		case <-newCtx.Done():
			return 0, errors.New("确认超时")
		}
	}
}

func (t *Wallet) BuildTx(
	privObj solana.PrivateKey,
	signers map[solana.PublicKey]*solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64, // 0 则使用默认
) (*solana.Transaction, error) {
	if len(instructions) == 0 {
		return nil, errors.Errorf("指令集为空")
	}

	userAddress := privObj.PublicKey()

	feeInstructions := make([]solana.Instruction, 0)
	if unitPrice != 0 {
		if unitLimit != 0 {
			feeInstructions = append(feeInstructions, computebudget.NewSetComputeUnitLimitInstruction(uint32(unitLimit)).Build())
		}
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

	if signers == nil {
		signers = map[solana.PublicKey]*solana.PrivateKey{
			userAddress: &privObj,
		}
	}
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			return signers[key]
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Sign error.")
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

func (t *Wallet) SendAndConfirmTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	skipPreflight bool,
	urls []string,
) (*rpc.GetParsedTransactionResult, error) {
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
					return nil, err
				}
				// t.logger.Error(err.Error())
			}
			getTransactionResult, err := t.rpcClient.GetParsedTransaction(
				ctx,
				tx.Signatures[0],
				&rpc.GetParsedTransactionOpts{
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: constant.MaxSupportedTransactionVersion_0,
				},
			)
			if err != nil {
				t.logger.Debug(err)
				confirmTimer.Reset(time.Second)
				continue
			}
			if getTransactionResult == nil {
				confirmTimer.Reset(time.Second)
				continue
			}

			if getTransactionResult.Meta.Err != nil {
				t.logger.InfoF("交易已确认[执行失败] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
				return getTransactionResult, errors.Errorf("<txid: %s> <err: %s>", tx.Signatures[0], go_format.ToString(getTransactionResult.Meta.Err))
			}
			t.logger.InfoF("交易已确认[执行成功] <%d>。<%s>", *getTransactionResult.BlockTime*1000, tx.Signatures[0].String())
			return getTransactionResult, nil
		case <-newCtx.Done():
			return nil, errors.New("确认超时")
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

func (t *Wallet) TransferSOL(
	from solana.PublicKey,
	to solana.PublicKey,
	amountWithDecimals uint64,
) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0)

	instructions = append(instructions,
		system.NewTransferInstruction(
			amountWithDecimals,
			from,
			to,
		).Build(),
	)

	return instructions, nil

}

func (t *Wallet) Balance(address solana.PublicKey) (uint64, error) {
	getAccountInfoResult, err := t.rpcClient.GetAccountInfo(context.Background(), address)
	if err != nil {
		return 0, errors.Wrap(err, "GetAccountInfo")
	}

	return getAccountInfoResult.Value.Lamports, nil

}

func (t *Wallet) DestroyTokenAccounts(
	userAddress solana.PublicKey,
	isForce bool, // 如果是 false，则跳过有余额的；如果是 true，则不管有多少余额，都直接销毁
	excludeTokenAddresses []solana.PublicKey, // isForce 为 true 时生效
) (
	instructions_ []solana.Instruction,
	closedTokenAccounts_ []solana.PublicKey, // 所有被关闭的 token account
	remainTokenAccounts_ []solana.PublicKey, // 所有剩下的没有被关闭的
	err_ error,
) {
	instructions := make([]solana.Instruction, 0)
	closedTokenAccounts := make([]solana.PublicKey, 0)
	remainTokenAccounts := make([]solana.PublicKey, 0)

	getTokenAccountsResult, err := t.rpcClient.GetTokenAccountsByOwner(
		context.Background(),
		userAddress,
		&rpc.GetTokenAccountsConfig{
			ProgramId: &solana.TokenProgramID,
		},
		&rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentConfirmed,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "")
	}
	if getTokenAccountsResult == nil {
		return nil, nil, nil, errors.Errorf("GetTokenAccountsByOwner failed.")
	}
	for _, tokenAccountInfo := range getTokenAccountsResult.Value {
		if len(closedTokenAccounts) > 30 {
			remainTokenAccounts = append(remainTokenAccounts, tokenAccountInfo.Pubkey)
			continue
		}

		var data associated_token_account.AssociatedTokenAccountDataType
		err = json.Unmarshal(tokenAccountInfo.Account.Data.GetRawJSON(), &data)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "")
		}
		// fmt.Printf("<%s> <tokenAddress: %s> <%s>\n", tokenAccountInfo.Pubkey.String(), data.Parsed.Info.Mint, data.Parsed.Info.TokenAmount.UIAmountString)

		if data.Parsed.Info.TokenAmount.UIAmount == 0 {
			instructions = append(instructions, token.NewCloseAccountInstruction(
				tokenAccountInfo.Pubkey,
				userAddress,
				userAddress,
				nil,
			).Build())
			closedTokenAccounts = append(closedTokenAccounts, tokenAccountInfo.Pubkey)
			continue
		}
		if !isForce {
			remainTokenAccounts = append(remainTokenAccounts, tokenAccountInfo.Pubkey)
			continue
		}
		if slices.Contains(excludeTokenAddresses, solana.MustPublicKeyFromBase58(data.Parsed.Info.Mint)) {
			remainTokenAccounts = append(remainTokenAccounts, tokenAccountInfo.Pubkey)
			continue
		}
		// 将余额 burn
		amountWithDecimals, err := strconv.ParseUint(data.Parsed.Info.TokenAmount.Amount, 10, 64)
		if err != nil {
			return nil, nil, nil, err
		}
		instructions = append(instructions,
			token.NewBurnInstruction(
				amountWithDecimals,
				tokenAccountInfo.Pubkey,
				solana.MustPublicKeyFromBase58(data.Parsed.Info.Mint),
				userAddress,
				nil,
			).Build(),
			token.NewCloseAccountInstruction(
				tokenAccountInfo.Pubkey,
				userAddress,
				userAddress,
				nil,
			).Build(),
		)
		closedTokenAccounts = append(closedTokenAccounts, tokenAccountInfo.Pubkey)
	}

	return instructions, closedTokenAccounts, remainTokenAccounts, nil
}
