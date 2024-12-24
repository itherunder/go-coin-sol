package go_coin_sol

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	constant "github.com/pefish/go-coin-sol/constant"
	i_logger "github.com/pefish/go-interface/i-logger"
)

type Wallet struct {
	logger    i_logger.ILogger
	rpcClient *rpc.Client
	wsClient  *ws.Client
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
	wsClient, err := ws.Connect(ctx, wssUrl)
	if err != nil {
		return nil, err
	}
	return &Wallet{
		logger:    logger,
		rpcClient: rpcClient,
		wsClient:  wsClient,
	}, nil
}

func (t *Wallet) RPCClient() *rpc.Client {
	return t.rpcClient
}

func (t *Wallet) WSClient() *ws.Client {
	return t.wsClient
}

func (t *Wallet) NewAddress() (address_ string, priv_ string) {
	account := solana.NewWallet()
	return account.PublicKey().String(), account.PrivateKey.String()
}

func (t *Wallet) SendTx(
	ctx context.Context,
	privObj solana.PrivateKey,
	latestBlockhash *solana.Hash,
	instructions []solana.Instruction,
	unitPrice uint64,
	unitLimit uint64,
	skipPreflight bool,
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
	t.logger.InfoF("交易构建成功。<%s>", tx.Signatures[0].String())

	newCtx, _ := context.WithTimeout(ctx, 60*time.Second)
	meta, timestamp, err := t.SendAndConfirmTransaction(newCtx, tx, skipPreflight)
	if err != nil {
		return nil, nil, 0, err
	}
	t.logger.InfoF("交易已确认。<%s>", tx.Signatures[0].String())

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
			return nil, err
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
		return nil, err
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
		return nil, err
	}
	return tx, nil
}

func (t *Wallet) SendAndConfirmTransaction(
	ctx context.Context,
	tx *solana.Transaction,
	skipPreflight bool,
) (
	meta_ *rpc.TransactionMeta,
	timestamp_ uint64,
	err_ error,
) {
	confirmTimer := time.NewTimer(0)
	for {
		select {
		case <-confirmTimer.C:
			_, err := t.rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight: skipPreflight,
			})
			if err != nil {
				t.logger.Error(err.Error())
				confirmTimer.Reset(500 * time.Millisecond)
				continue
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
			return getTransactionResult.Meta, uint64(*getTransactionResult.BlockTime * 1000), nil
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		}
	}

}
