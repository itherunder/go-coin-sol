package raydium

import (
	"encoding/hex"
	"strconv"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	associated_token_account "github.com/pefish/go-coin-sol/program/associated-token-account"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	"github.com/pefish/go-coin-sol/program/raydium/instruction"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

func GetSwapInstructions(
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	raydiumSwapKeys raydium_type_.RaydiumSwapKeys,
	isClose bool,
	solReserveWithDecimals uint64,
	tokenReserveWithDecimals uint64,
	slippage int64,
) ([]solana.Instruction, error) {
	if slippage == 0 {
		slippage = 50 // 0.5%
	}
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
		if slippage == -1 {
			return nil, errors.New("购买必须设置滑点")
		}
		maxCostSolAmountWithDecimals := uint64(
			float64(slippage+10000) * 1.005 * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
		) // raydium 收取 0.25% 交易手续费

		swapInstruction, err := instruction.NewBuyBaseOutInstruction(
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmountWithDecimals,
			maxCostSolAmountWithDecimals,
			raydiumSwapKeys,
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
		if slippage != -1 {
			// 应该收到的 sol 数量
			minReceiveSolAmountWithDecimals = uint64(
				0.995 * float64(10000-slippage) * float64(solReserveWithDecimals) * float64(tokenAmountWithDecimals) / float64(tokenReserveWithDecimals) / 10000,
			)
		}

		swapInstruction, err := instruction.NewSellBaseInInstruction(
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmountWithDecimals,
			minReceiveSolAmountWithDecimals,
			raydiumSwapKeys,
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

func GetReserves(
	rpcClient *rpc.Client,
	poolCoinTokenAccount solana.PublicKey,
	poolPcTokenAccount solana.PublicKey,
) (
	reserveSol_ *type_.TokenAmountInfo,
	reserveToken_ *type_.TokenAmountInfo,
	err_ error,
) {
	datas, err := associated_token_account.GetAssociatedTokenAccountDatas(
		rpcClient,
		[]solana.PublicKey{
			poolCoinTokenAccount,
			poolPcTokenAccount,
		},
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "<poolCoinTokenAccount: %s> <poolPcTokenAccount: %s>", poolCoinTokenAccount, poolPcTokenAccount)
	}
	if datas[0] == nil || datas[1] == nil {
		return nil, nil, errors.New("raydium 账户没查到信息")
	}
	solAmountWithDecimals, err := strconv.ParseUint(datas[0].Parsed.Info.TokenAmount.Amount, 10, 64)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "<amount: %s>", datas[0].Parsed.Info.TokenAmount.Amount)
	}
	tokenAmountWithDecimals, err := strconv.ParseUint(datas[1].Parsed.Info.TokenAmount.Amount, 10, 64)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "<amount: %s>", datas[1].Parsed.Info.TokenAmount.Amount)
	}
	return &type_.TokenAmountInfo{
			AmountWithDecimals: solAmountWithDecimals,
			Decimals:           datas[0].Parsed.Info.TokenAmount.Decimals,
		},
		&type_.TokenAmountInfo{
			AmountWithDecimals: tokenAmountWithDecimals,
			Decimals:           datas[1].Parsed.Info.TokenAmount.Decimals,
		},
		nil
}

func ParseSwapTx(meta *rpc.TransactionMeta, transaction *solana.Transaction) (*raydium_type_.SwapTxDataType, error) {
	swaps := make([]*raydium_type_.SwapDataType, 0)

	accountKeys := transaction.Message.AccountKeys
	if meta.LoadedAddresses.Writable != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.Writable...)
	}
	if meta.LoadedAddresses.ReadOnly != nil {
		accountKeys = append(accountKeys, meta.LoadedAddresses.ReadOnly...)
	}

	allInstructions := make([]solana.CompiledInstruction, 0)
	for index, instruction := range transaction.Message.Instructions {
		allInstructions = append(allInstructions, instruction)
		innerInstructions := util.FindInnerInstructions(meta, uint64(index))
		if innerInstructions == nil {
			continue
		}
		allInstructions = append(allInstructions, innerInstructions...)
	}

	for index, instruction := range allInstructions {
		programPKey := accountKeys[instruction.ProgramIDIndex]
		if !programPKey.Equals(raydium_constant.Raydium_Liquidity_Pool_V4) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:2]
		if methodId != "0b" && methodId != "09" {
			continue
		}

		poolCoinTokenAccount := accountKeys[instruction.Accounts[len(instruction.Accounts)-13]] // wsol
		poolPCTokenAccount := accountKeys[instruction.Accounts[len(instruction.Accounts)-12]]   // token

		transfer1Instruction := allInstructions[index+1]
		transfer2Instruction := allInstructions[index+2]

		var transfer1InstructionData struct {
			Discriminator uint8  `json:"discriminator"`
			Amount        uint64 `json:"amount"`
		}
		err := bin.NewBorshDecoder(transfer1Instruction.Data).Decode(&transfer1InstructionData)
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s> <data: %s>", transaction.Signatures[0].String(), transfer1Instruction.Data.String())
		}

		var transfer2InstructionData struct {
			Discriminator uint8  `json:"discriminator"`
			Amount        uint64 `json:"amount"`
		}
		err = bin.NewBorshDecoder(transfer2Instruction.Data).Decode(&transfer2InstructionData)
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s> <data: %s>", transaction.Signatures[0].String(), transfer2Instruction.Data.String())
		}
		var swapType type_.SwapType
		var solAmountWithDecimals uint64
		var tokenAmountWithDecimals uint64
		if accountKeys[transfer1Instruction.Accounts[1]].Equals(poolCoinTokenAccount) {
			swapType = type_.SwapType_Buy
			solAmountWithDecimals = transfer1InstructionData.Amount
			tokenAmountWithDecimals = transfer2InstructionData.Amount
		} else {
			swapType = type_.SwapType_Sell
			solAmountWithDecimals = transfer2InstructionData.Amount
			tokenAmountWithDecimals = transfer1InstructionData.Amount
		}

		userAddress := accountKeys[0]
		// 得到 token address
		var tokenAddress solana.PublicKey
		// fmt.Println(userTokenAssociatedAccount.String(), userAddress)
		var tokenBalanceInfo *rpc.TokenBalance
		for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(raydium_constant.Raydium_Authority_V4) &&
				accountKeys[tokenBalanceInfo_.AccountIndex].Equals(poolPCTokenAccount) {
				tokenBalanceInfo = &tokenBalanceInfo_
				break
			}
		}
		if tokenBalanceInfo == nil {
			return nil, errors.Errorf("没有找到 token 相关的 balance info. <txid: %s>", transaction.Signatures[0].String())
		}
		tokenAddress = tokenBalanceInfo.Mint
		// 得到交易前后 token 的余额
		userTokenAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
			userAddress,
			tokenAddress,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "<userAddress: %s> <tokenAddress: %s>", userAddress, tokenAddress)
		}
		// fmt.Println(userTokenAssociatedAccount.String())
		beforeUserTokenBalanceWithDecimals := uint64(0)
		userTokenBalanceWithDecimals := uint64(0)
		for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(userAddress) &&
				accountKeys[tokenBalanceInfo_.AccountIndex].Equals(userTokenAssociatedAccount) {
				beforeUserTokenBalanceWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
				break
			}
		}
		for _, tokenBalanceInfo_ := range meta.PostTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(userAddress) &&
				accountKeys[tokenBalanceInfo_.AccountIndex].Equals(userTokenAssociatedAccount) {
				userTokenBalanceWithDecimals, _ = strconv.ParseUint(tokenBalanceInfo_.UiTokenAmount.Amount, 10, 64)
				break
			}
		}

		swaps = append(swaps, &raydium_type_.SwapDataType{
			TokenAddress:                       tokenAddress,
			SOLAmountWithDecimals:              solAmountWithDecimals,
			TokenAmountWithDecimals:            tokenAmountWithDecimals,
			Type:                               swapType,
			UserAddress:                        userAddress,
			BeforeUserBalanceWithDecimals:      meta.PreBalances[0],
			UserBalanceWithDecimals:            meta.PostBalances[0],
			BeforeUserTokenBalanceWithDecimals: beforeUserTokenBalanceWithDecimals,
			UserTokenBalanceWithDecimals:       userTokenBalanceWithDecimals,
			RaydiumSwapKeys: raydium_type_.RaydiumSwapKeys{
				AmmAddress:                   accountKeys[instruction.Accounts[1]],
				AmmOpenOrdersAddress:         &accountKeys[instruction.Accounts[3]],
				AmmTargetOrdersAddress:       &accountKeys[instruction.Accounts[len(instruction.Accounts)-14]],
				PoolCoinTokenAccountAddress:  poolCoinTokenAccount,
				PoolPcTokenAccountAddress:    poolPCTokenAccount,
				SerumProgramAddress:          &accountKeys[instruction.Accounts[len(instruction.Accounts)-11]],
				SerumMarketAddress:           &accountKeys[instruction.Accounts[len(instruction.Accounts)-10]],
				SerumBidsAddress:             &accountKeys[instruction.Accounts[len(instruction.Accounts)-9]],
				SerumAsksAddress:             &accountKeys[instruction.Accounts[len(instruction.Accounts)-8]],
				SerumEventQueueAddress:       &accountKeys[instruction.Accounts[len(instruction.Accounts)-7]],
				SerumCoinVaultAccountAddress: &accountKeys[instruction.Accounts[len(instruction.Accounts)-6]],
				SerumPcVaultAccountAddress:   &accountKeys[instruction.Accounts[len(instruction.Accounts)-5]],
				SerumVaultSignerAddress:      &accountKeys[instruction.Accounts[len(instruction.Accounts)-4]],
			},
		})
	}

	feeInfo, err := util.GetFeeInfoFromTx(meta, transaction)
	if err != nil {
		return nil, err
	}

	return &raydium_type_.SwapTxDataType{
		TxId:    transaction.Signatures[0].String(),
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}
