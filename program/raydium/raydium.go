package raydium

import (
	"encoding/hex"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	associated_token_account "github.com/pefish/go-coin-sol/program/associated-token-account"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	raydium_constant "github.com/pefish/go-coin-sol/program/raydium/constant"
	"github.com/pefish/go-coin-sol/program/raydium/instruction"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	go_decimal "github.com/pefish/go-decimal"
)

func GetSwapInstructions(
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmount type_.TokenAmountInfo,
	raydiumSwapKeys raydium_type_.RaydiumSwapKeys,
	isClose bool,
	solReserve string,
	tokenReserve string,
	slippage uint64,
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
		return nil, err
	}

	userTokenAssociatedAccount, _, err := solana.FindAssociatedTokenAddress(
		userAddress,
		tokenAddress,
	)
	if err != nil {
		return nil, err
	}

	if swapType == type_.SwapType_Buy {
		// 应该花费的 sol 数量
		shouldCostSolAmount := go_decimal.Decimal.MustStart(solReserve).MustMulti(tokenAmount.Amount).MustDiv(tokenReserve).MustMultiForString(1.005) // raydium 收取 0.25% 交易手续费
		// 最大多花 sol 的数量
		maxMoreSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustMulti(slippage).MustDivForString(10000)
		maxCostSolAmount := go_decimal.Decimal.MustStart(shouldCostSolAmount).MustAdd(maxMoreSolAmount).RoundDownForString(constant.SOL_Decimals)

		swapInstruction, err := instruction.NewBuyBaseOutInstruction(
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmount,
			maxCostSolAmount,
			raydiumSwapKeys,
		)
		if err != nil {
			return nil, err
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
				go_decimal.Decimal.MustStart(maxCostSolAmount).MustShiftedBy(constant.SOL_Decimals).RoundDown(0).MustEndForUint64(),
				userAddress,
				userWSOLAssociatedAccount,
			).Build(),
			token.NewSyncNativeInstruction(userWSOLAssociatedAccount).Build(),
			swapInstruction,
		)
	} else {
		// 应该收到的 sol 数量
		shouldReceiveSolAmount := go_decimal.Decimal.MustStart(solReserve).MustMulti(tokenAmount.Amount).MustDiv(tokenReserve).MustMultiForString(0.995)
		// 最大少收到 sol 的数量
		maxLessSolAmount := go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustMulti(slippage).MustDivForString(10000)
		minReceiveSolAmount := go_decimal.Decimal.MustStart(shouldReceiveSolAmount).MustSub(maxLessSolAmount).RoundDownForString(constant.SOL_Decimals)

		swapInstruction, err := instruction.NewSellBaseInInstruction(
			userAddress,
			tokenAddress,
			userWSOLAssociatedAccount,
			userTokenAssociatedAccount,
			tokenAmount,
			minReceiveSolAmount,
			raydiumSwapKeys,
		)
		if err != nil {
			return nil, err
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
		)

		if isClose {
			instructions = append(
				instructions,
				token.NewCloseAccountInstruction(
					userWSOLAssociatedAccount,
					userAddress,
					userAddress,
					nil,
				).Build(),
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
	reserveSolAmount_ string,
	reserveTokenAmount_ string,
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
		return "", "", err
	}
	return datas[0].Parsed.Info.TokenAmount.UIAmountString,
		datas[1].Parsed.Info.TokenAmount.UIAmountString,
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

		poolCoinTokenAccount := accountKeys[instruction.Accounts[len(instruction.Accounts)-13]]

		transfer1Instruction := allInstructions[index+1]
		transfer2Instruction := allInstructions[index+2]

		var transfer1InstructionData struct {
			Discriminator uint8  `json:"discriminator"`
			Amount        uint64 `json:"amount"`
		}
		err := bin.NewBorshDecoder(transfer1Instruction.Data).Decode(&transfer1InstructionData)
		if err != nil {
			return nil, err
		}

		var transfer2InstructionData struct {
			Discriminator uint8  `json:"discriminator"`
			Amount        uint64 `json:"amount"`
		}
		err = bin.NewBorshDecoder(transfer2Instruction.Data).Decode(&transfer2InstructionData)
		if err != nil {
			return nil, err
		}
		var swapType type_.SwapType
		var solAmount string
		var tokenAmountWithDecimals uint64
		if accountKeys[transfer1Instruction.Accounts[1]].Equals(poolCoinTokenAccount) {
			swapType = type_.SwapType_Buy
			solAmount = go_decimal.Decimal.MustStart(transfer1InstructionData.Amount).MustUnShiftedBy(constant.SOL_Decimals).EndForString()
			tokenAmountWithDecimals = transfer2InstructionData.Amount
		} else {
			swapType = type_.SwapType_Sell
			solAmount = go_decimal.Decimal.MustStart(transfer2InstructionData.Amount).MustUnShiftedBy(constant.SOL_Decimals).EndForString()
			tokenAmountWithDecimals = transfer1InstructionData.Amount
		}
		swaps = append(swaps, &raydium_type_.SwapDataType{
			SOLAmount:               solAmount,
			TokenAmountWithDecimals: tokenAmountWithDecimals,
			Type:                    swapType,
			UserAddress:             accountKeys[instruction.Accounts[16]],
			RaydiumSwapKeys: raydium_type_.RaydiumSwapKeys{
				AmmAddress:                   accountKeys[instruction.Accounts[1]],
				AmmOpenOrdersAddress:         &accountKeys[instruction.Accounts[3]],
				AmmTargetOrdersAddress:       &accountKeys[instruction.Accounts[len(instruction.Accounts)-14]],
				PoolCoinTokenAccountAddress:  poolCoinTokenAccount,
				PoolPcTokenAccountAddress:    accountKeys[instruction.Accounts[len(instruction.Accounts)-12]],
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
