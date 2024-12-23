package raydium

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pefish/go-coin-sol/constant"
	associated_token_account "github.com/pefish/go-coin-sol/program/associated-token-account"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	"github.com/pefish/go-coin-sol/program/raydium/instruction"
	raydium_type_ "github.com/pefish/go-coin-sol/program/raydium/type"
	type_ "github.com/pefish/go-coin-sol/type"
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
