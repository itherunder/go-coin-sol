package raydium_clmm

// Concentrated Liquidity (CLMM)

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	"github.com/pefish/go-coin-sol/program/raydium-clmm/instruction"
	raydium_clmm_type "github.com/pefish/go-coin-sol/program/raydium-clmm/type"
	type_ "github.com/pefish/go-coin-sol/type"
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

func GetReserves(
	rpcClient *rpc.Client,
	poolIdAddress solana.PublicKey,
) (
	reserveSol_ *type_.TokenAmountInfo,
	reserveToken_ *type_.TokenAmountInfo,
	err_ error,
) {
	poolInfo, err := GetPoolInfo(rpcClient, poolIdAddress)
	if err != nil {
		return nil, nil, err
	}
	var reserveSolAmountInfo type_.TokenAmountInfo
	var reserveTokenAmountInfo type_.TokenAmountInfo
	if poolInfo.TokenMint0.Equals(solana.SolMint) {
		reserveSolAmountInfo.AmountWithDecimals = poolInfo.SwapInAmountToken0.BigInt().Uint64()
		reserveSolAmountInfo.Decimals = uint64(poolInfo.MintDecimals0)

		reserveTokenAmountInfo.AmountWithDecimals = poolInfo.SwapInAmountToken1.BigInt().Uint64()
		reserveTokenAmountInfo.Decimals = uint64(poolInfo.MintDecimals1)
	} else {
		reserveSolAmountInfo.AmountWithDecimals = poolInfo.SwapInAmountToken1.BigInt().Uint64()
		reserveSolAmountInfo.Decimals = uint64(poolInfo.MintDecimals1)

		reserveTokenAmountInfo.AmountWithDecimals = poolInfo.SwapInAmountToken0.BigInt().Uint64()
		reserveTokenAmountInfo.Decimals = uint64(poolInfo.MintDecimals0)
	}

	return &reserveSolAmountInfo, &reserveTokenAmountInfo, nil
}

func GetSwapInstructions(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	swapKeys raydium_clmm_type.SwapKeys,
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
