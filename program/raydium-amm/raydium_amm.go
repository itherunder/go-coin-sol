package raydium_amm

// Legacy AMM v4

import (
	"encoding/hex"
	"strconv"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	associated_token_account_instruction "github.com/pefish/go-coin-sol/program/associated-token-account/instruction"
	pumpfun_constant "github.com/pefish/go-coin-sol/program/pumpfun/constant"
	raydium_amm_constant "github.com/pefish/go-coin-sol/program/raydium-amm/constant"
	"github.com/pefish/go-coin-sol/program/raydium-amm/instruction"
	raydium_amm_type "github.com/pefish/go-coin-sol/program/raydium-amm/type"
	type_ "github.com/pefish/go-coin-sol/type"
	"github.com/pefish/go-coin-sol/util"
	"github.com/pkg/errors"
)

func GetSwapInstructions(
	network rpc.Cluster,
	userAddress solana.PublicKey,
	swapType type_.SwapType,
	tokenAddress solana.PublicKey,
	tokenAmountWithDecimals uint64,
	raydiumSwapKeys raydium_amm_type.SwapKeys,
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
		) // raydium 收取 0.25% 交易手续费

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
		if !instruction.ProgramId.Equals(raydium_amm_constant.Raydium_Liquidity_Pool_V4[network]) {
			continue
		}
		methodId := hex.EncodeToString(instruction.Data)[:2]
		if methodId != "0b" && methodId != "09" {
			continue
		}

		poolCoinTokenAccount := instruction.Accounts[len(instruction.Accounts)-13]
		poolPCTokenAccount := instruction.Accounts[len(instruction.Accounts)-12]

		transfer1Data, err := util.DecodeTransferInstruction(allInstructions[index+1])
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s>", txId)
		}
		transfer2Data, err := util.DecodeTransferInstruction(allInstructions[index+2])
		if err != nil {
			return nil, errors.Wrapf(err, "<txid: %s>", txId)
		}
		inputAmountWithDecimals := transfer1Data.AmountWithDecimals
		outputAmountWithDecimals := transfer2Data.AmountWithDecimals

		var coinAddress solana.PublicKey
		var pcAddress solana.PublicKey
		var coinDecimals uint64
		var pcDecimals uint64
		for _, tokenBalanceInfo_ := range meta.PreTokenBalances {
			if tokenBalanceInfo_.Owner.Equals(raydium_amm_constant.Raydium_Authority_V4[network]) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(poolCoinTokenAccount) {
				coinAddress = tokenBalanceInfo_.Mint
				coinDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
			}
			if tokenBalanceInfo_.Owner.Equals(raydium_amm_constant.Raydium_Authority_V4[network]) &&
				transaction.Message.AccountKeys[tokenBalanceInfo_.AccountIndex].PublicKey.Equals(poolPCTokenAccount) {
				pcAddress = tokenBalanceInfo_.Mint
				pcDecimals = uint64(tokenBalanceInfo_.UiTokenAmount.Decimals)
			}
		}

		var inputVault solana.PublicKey
		var outputVault solana.PublicKey
		var inputAddress solana.PublicKey
		var outputAddress solana.PublicKey
		var inputDecimals uint64
		var outputDecimals uint64
		if transfer1Data.Destination.Equals(poolCoinTokenAccount) {
			// coin is input
			inputVault = poolCoinTokenAccount
			outputVault = poolPCTokenAccount
			inputAddress = coinAddress
			outputAddress = pcAddress
			inputDecimals = coinDecimals
			outputDecimals = pcDecimals
		} else {
			outputVault = poolCoinTokenAccount
			inputVault = poolPCTokenAccount
			inputAddress = pcAddress
			outputAddress = coinAddress
			inputDecimals = pcDecimals
			outputDecimals = coinDecimals
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
			InputAddress:             inputAddress,
			OutputAddress:            outputAddress,
			InputAmountWithDecimals:  inputAmountWithDecimals,
			InputDecimals:            inputDecimals,
			OutputAmountWithDecimals: outputAmountWithDecimals,
			OutputDecimals:           outputDecimals,
			UserAddress:              userAddress,
			ParsedKeys: &raydium_amm_type.SwapKeys{
				AmmAddress:                  instruction.Accounts[1],
				PoolCoinTokenAccountAddress: poolCoinTokenAccount,
				PoolPcTokenAccountAddress:   poolPCTokenAccount,
				CoinMint:                    coinAddress,
				PCMint:                      pcAddress,
				Vaults: map[solana.PublicKey]solana.PublicKey{
					pcAddress:   poolPCTokenAccount,
					coinAddress: poolCoinTokenAccount,
				},
			},
			PairAddress:               instruction.Accounts[1],
			InputVault:                inputVault,
			OutputVault:               outputVault,
			ReserveInputWithDecimals:  reserveInputWithDecimals,
			ReserveOutputWithDecimals: reserveOutputWithDecimals,
			Keys:                      instruction.Accounts,
			MethodId:                  methodId,
			Program:                   raydium_amm_constant.Raydium_Authority_V4[network],
		})
	}

	return &type_.SwapTxDataType{
		TxId:    txId,
		Swaps:   swaps,
		FeeInfo: feeInfo,
	}, nil
}

func ParseAddLiqTxByParsedTx(
	network rpc.Cluster,
	meta *rpc.ParsedTransactionMeta,
	parsedTransaction *rpc.ParsedTransaction,
) (*raydium_amm_type.AddLiqTxDataType, error) {
	if !parsedTransaction.Message.AccountKeys[0].PublicKey.Equals(pumpfun_constant.Pumpfun_Raydium_Migration) {
		return nil, nil
	}
	for _, parsedInstruction := range parsedTransaction.Message.Instructions {
		if !parsedInstruction.ProgramId.Equals(raydium_amm_constant.Raydium_Liquidity_Pool_V4[network]) {
			continue
		}
		if hex.EncodeToString(parsedInstruction.Data)[:2] != "01" {
			continue
		}
		var params struct {
			Discriminator  uint8  `json:"discriminator"`
			Nonce          uint8  `json:"nonce"`
			OpenTime       uint64 `json:"openTime"`
			InitPcAmount   uint64 `json:"initPcAmount"`
			InitCoinAmount uint64 `json:"initCoinAmount"`
		}
		err := bin.NewBorshDecoder(parsedInstruction.Data).Decode(&params)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		feeInfo, err := util.GetFeeInfoFromParsedTx(meta, parsedTransaction)
		if err != nil {
			return nil, err
		}

		coinIsSOL := parsedInstruction.Accounts[8].Equals(solana.SolMint)
		var solVault solana.PublicKey
		var tokenVault solana.PublicKey
		var tokenAddress solana.PublicKey
		if coinIsSOL {
			tokenAddress = parsedInstruction.Accounts[9]
			solVault = parsedInstruction.Accounts[10]
			tokenVault = parsedInstruction.Accounts[11]
		} else {
			tokenAddress = parsedInstruction.Accounts[8]
			solVault = parsedInstruction.Accounts[11]
			tokenVault = parsedInstruction.Accounts[10]
		}

		return &raydium_amm_type.AddLiqTxDataType{
			TxId:                        parsedTransaction.Signatures[0].String(),
			TokenAddress:                tokenAddress,
			AMMAddress:                  parsedInstruction.Accounts[4],
			PoolCoinTokenAccount:        parsedInstruction.Accounts[10],
			PoolPcTokenAccount:          parsedInstruction.Accounts[11],
			InitSOLAmountWithDecimals:   params.InitCoinAmount,
			InitTokenAmountWithDecimals: params.InitPcAmount,
			PairAddress:                 parsedInstruction.Accounts[4],
			SOLVault:                    solVault,
			TokenVault:                  tokenVault,
			FeeInfo:                     feeInfo,
		}, nil

	}

	return nil, nil
}
