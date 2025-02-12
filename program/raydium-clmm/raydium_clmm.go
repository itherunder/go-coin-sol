package raydium_clmm

// Concentrated Liquidity (CLMM)

import (
	"context"
	"encoding/hex"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/discriminator"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	raydium_clmm_constant "github.com/pefish/go-coin-sol/program/raydium-clmm/constant"
	"github.com/pefish/go-coin-sol/program/raydium-clmm/instruction"
	raydium_clmm_type "github.com/pefish/go-coin-sol/program/raydium-clmm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

func GetPoolInfo(
	rpcClient *rpc.Client,
	poolIdAddress solana.PublicKey,
) (*raydium_clmm_type.PoolInfo, error) {
	var data raydium_clmm_type.PoolInfo
	err := rpcClient.GetAccountDataBorshInto(context.Background(), poolIdAddress, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "<poolIdAddress: %s>", poolIdAddress)
	}

	return &data, nil
}

func GetSwapInstructions(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	swapKeys raydium_clmm_type.SwapV2Keys,
	isClose bool,
	solReserveWithDecimals uint64,
	tokenReserveWithDecimals uint64,
	slippage uint64, // 0 代表不设置滑点
) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0)

	userWSOLAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		solana.SolMint,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s>", userAddress)
	}

	userTokenAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		tokenAddress,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, tokenAddress)
	}

	if swapType == type_.SwapType_Buy {
		if slippage == 0 {
			return nil, errors.New("购买必须设置滑点")
		}
		maxCostSolAmountWithDecimals := uint64(
			float64(slippage+10000) * 1.005 * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
		)

		if maxCostSolAmountWithDecimals == 0 {
			return nil, errors.New("购买数量太小")
		}

		swapInstruction, err := instruction.NewBuyBaseOutInstruction(
			network,
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmountWithDecimals,
			maxCostSolAmountWithDecimals,
			swapKeys,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		instructions = append(
			instructions,
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userWSOLAssociatedAccount,
				userAddress,
				solana.SolMint,
			),
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userTokenAssociatedAccount,
				userAddress,
				tokenAddress,
			),
			system.NewTransferInstruction(
				maxCostSolAmountWithDecimals,
				userAddress,
				userWSOLAssociatedAccount,
			).Build(),
			token.NewSyncNativeInstruction(userWSOLAssociatedAccount).Build(),
			swapInstruction,
		)
	} else {
		minReceiveSolAmountWithDecimals := uint64(0)
		if slippage != 0 {
			// 应该收到的 sol 数量
			minReceiveSolAmountWithDecimals = uint64(
				0.995 * float64(10000-slippage) * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
			)
		}

		swapInstruction, err := instruction.NewSellBaseInInstruction(
			network,
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmountWithDecimals,
			minReceiveSolAmountWithDecimals,
			swapKeys,
		)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		instructions = append(
			instructions,
			associated_token_account_instruction.NewCreateIdempotentInstruction(
				userAddress,
				userWSOLAssociatedAccount,
				userAddress,
				solana.SolMint,
			),
			swapInstruction,
			token.NewCloseAccountInstruction(
				userWSOLAssociatedAccount,
				userAddress,
				userAddress,
				nil,
			).Build(),
		)

		if isClose {
			instructions = append(
				instructions,
				token.NewCloseAccountInstruction(
					userTokenAssociatedAccount,
					userAddress,
					userAddress,
					nil,
				).Build(),
			)
		}
	}

	return instructions, nil
}

func ParseSwapTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	transaction *rpc.ParsedTransaction,
) (*type_.SwapTxDataType, error) {
	txId := transaction.Signatures[0].String()
	feeInfo, err := util.GetFeeInfoFromParsedTx(meta, transaction)
	if err != nil {
		return nil, errors.Wrapf(err, "<txid: %s>", txId)
	}
	swaps := make([]*type_.SwapDataType, 0)
	if meta.Err != nil {
		return &type_.SwapTxDataType{
			TxId:    txId,
			Swaps:   swaps,
			FeeInfo: feeInfo,
		}, nil
	}

	allInstructions := make([]*rpc.ParsedInstruction, 0)
	for index, instruction := range transaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructionsFromParsedMeta(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	for index, instruction := range allInstructions {
		if !instruction.ProgramId.Equals(raydium_clmm_constant.Raydium_Concentrated_Liquidity[network]) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:16]
		inputVault := instruction.Accounts[5]
		outputVault := instruction.Accounts[6]
		pairAddress := instruction.Accounts[2]
		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		var parsedKeys interface{}
		var inputAmountWithDecimals uint64
		var outputAmountWithDecimals uint64

		if methodId == discriminator.GetDiscriminator("global", "swap_v2") {
			inputAddress = instruction.Accounts[11]
			outputAddress = instruction.Accounts[12]
			parsedKeys = &raydium_clmm_type.SwapV2Keys{
				AmmConfig:   instruction.Accounts[1],
				PairAddress: instruction.Accounts[2],
				Vaults: map[solana.PublicKey]solana.PublicKey{
					inputAddress:  inputVault,
					outputAddress: outputVault,
				},
				ObservationState: instruction.Accounts[7],
				ExBitmapAccount:  instruction.Accounts[13],
				RemainAccounts:   instruction.Accounts[14:],
			}
			transfer1Data, err := util.DecodeTransferCheckedInstruction(allInstructions[index+1])
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}
			transfer2Data, err := util.DecodeTransferCheckedInstruction(allInstructions[index+2])
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}
			inputAmountWithDecimals = transfer1Data.AmountWithDecimals
			outputAmountWithDecimals = transfer2Data.AmountWithDecimals
		} else if methodId == discriminator.GetDiscriminator("global", "swap") {
			for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
				if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
					transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(inputVault) {
					inputAddress = tokenBalanceInfo_.Mint
				}
				if tokenBalanceInfo_.Owner.Equals(pairAddress) &&
					transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(outputVault) {
					outputAddress = tokenBalanceInfo_.Mint
				}
			}

			parsedKeys = &raydium_clmm_type.SwapKeys{
				AmmConfig:   instruction.Accounts[1],
				PairAddress: instruction.Accounts[2],
				Vaults: map[solana.PublicKey]solana.PublicKey{
					inputAddress:  inputVault,
					outputAddress: outputVault,
				},
				ObservationState: instruction.Accounts[7],
				TickArrayAccount: instruction.Accounts[11],
				RemainAccounts:   instruction.Accounts[12:],
			}

			transfer1Data, err := util.DecodeTransferInstruction(allInstructions[index+1])
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}
			transfer2Data, err := util.DecodeTransferInstruction(allInstructions[index+2])
			if err != nil {
				return nil, errors.Wrapf(err, "<txid: %s>", txId)
			}
			inputAmountWithDecimals = transfer1Data.AmountWithDecimals
			outputAmountWithDecimals = transfer2Data.AmountWithDecimals
		} else {
			continue
		}

		userAddress := transaction.Message.AccountKeys[0].PublicKey

		var reserveInputWithDecimals uint64
		var reserveOutputWithDecimals uint64
		for _, tokenBalanceInfo_ := range meta.PostTokenBalances {
			if transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(inputVault) {
				reserveInputWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
			if transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(outputVault) {
				reserveOutputWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
			}
		}

		swaps = append(swaps, &type_.SwapDataType{
			InputAddress:              inputAddress,
			OutputAddress:             outputAddress,
			InputAmountWithDecimals:   inputAmountWithDecimals,
			OutputAmountWithDecimals:  outputAmountWithDecimals,
			UserAddress:               userAddress,
			PairAddress:               pairAddress,
			InputVault:                inputVault,
			OutputVault:               outputVault,
			ReserveInputWithDecimals:  reserveInputWithDecimals,
			ReserveOutputWithDecimals: reserveOutputWithDecimals,
			ParsedKeys:                parsedKeys,
			Keys:                      instruction.Accounts,
			MethodId:                  methodId,
			Program:                   raydium_clmm_constant.Raydium_Concentrated_Liquidity[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
